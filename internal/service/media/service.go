package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"image"
	// Supported image formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/streamio"

	"github.com/Karaoke-Manager/karman/internal/entity"

	// AV Libraries.
	"github.com/abema/go-mp4"
	"github.com/tcolgate/mp3"
)

// Service provides an interface for working with media files in Karman.
// An implementation of the Service interface implements the core logic associated with these files.
type Service interface {
	// StoreFile creates a new model.File and writes the data provided by r into the file.
	// This method should update known file metadata fields during the upload.
	// Depending on the media type implementations should analyze the file set type-specific metadata as well.
	//
	// If an error occurs r may have been partially consumed.
	// If any bytes have been persisted, this method must return a valid model.File that is able to identify the (potentially partial) data.
	// If the file has not been stored successfully, an error is returned.
	StoreFile(ctx context.Context, mediaType mediatype.MediaType, r io.Reader) (*model.File, error)

	// OpenFile creates a reader that can be used to read the file.
	// It is the caller's responsibility to close the reader when done.
	OpenFile(ctx context.Context, file *model.File) (io.ReadCloser, error)
}

// NewService creates a new Service instance using the supplied db and store.
// The default implementation will store media files in the store as well as in the DB.
// For each media file there will be an entry in the DB, the actual data however lives in the store.
func NewService(db *gorm.DB, store Store) Service {
	return &service{db, store}
}

// service is the default Service implementation.
type service struct {
	db    *gorm.DB
	store Store
}

// StoreFile creates a new entity.File in the database and then saves the data from r into the store.
// Supported media types are analyzed on the fly.
func (s *service) StoreFile(ctx context.Context, mediaType mediatype.MediaType, r io.Reader) (*model.File, error) {
	var err error
	file := entity.File{Type: mediaType}
	// We save the file here and at the end of the method to make sure that even
	// for half-written files an entry in the DB exists.
	// This makes finding orphaned files much easier.
	if err = s.db.WithContext(ctx).Save(&file).Error; err != nil {
		return nil, err
	}

	var w io.WriteCloser
	if w, err = s.store.CreateFile(ctx, file.Type, file.UUID); err != nil {
		return nil, err
	}
	defer func() {
		cErr := w.Close()
		if err == nil {
			err = cErr
		}
	}()

	wg := sync.WaitGroup{}
	pr, pw := io.Pipe()
	r = io.TeeReader(r, pw)

	wg.Add(1)
	// Analyze data in parallel to writing
	go func() {
		defer wg.Done()
		if err := fullAnalyzeFile(pr, mediaType, &file); err != nil {
			// TODO: Log unexpected error
			return
		}
	}()

	_, err = io.Copy(w, r)
	_ = pw.Close() // make sure that the goroutine stops
	wg.Wait()
	if err != nil {
		// This probably indicates that the request was cancelled or a network error occurred
		// Let background job do the cleanup
		// TODO: Log error
		return nil, err
	}

	if err = s.db.WithContext(ctx).Save(&file).Error; err != nil {
		return nil, err
	}
	return file.ToModel(), err
}

// fullAnalyzeFile reads the complete data from r and updates file to reflect its contents.
// Analysis includes fields like file size and checksum but
// also performs content-specific analysis for images, video and audio files.
func fullAnalyzeFile(r io.Reader, mediaType mediatype.MediaType, file *entity.File) error {
	var size int64
	r = streamio.CountBytes(r, &size)
	h := sha256.New()
	r = io.TeeReader(r, h)

	switch mediaType.Type() {
	case "image":
		_ = analyzeImage(r, mediaType, file)
	case "audio":
		_ = analyzeAudio(r, mediaType, file)
	case "video":
		_ = analyzeVideo(r, mediaType, file)
	}
	// TODO: Log unknown formats and other errors

	// Read remaining bytes for accurate size and hash.
	// We also want to make sure to read r until the end.
	if _, err := io.Copy(io.Discard, r); err != nil {
		return err
	}
	file.Size = size
	file.Checksum = h.Sum(nil)
	return nil
}

// analyzeImage sets image-specific metadata on file.
func analyzeImage(r io.Reader, mediaType mediatype.MediaType, file *entity.File) error {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return err
	}
	file.Width = cfg.Width
	file.Height = cfg.Height
	return nil
}

// analyzeAudio sets audio-specific metadata on file.
func analyzeAudio(r io.Reader, mediaType mediatype.MediaType, file *entity.File) error {
	switch mediaType.Subtype() {
	case "mpeg", "mpeg3", "x-mpeg-3", "mp3":
		duration := time.Duration(0)
		d := mp3.NewDecoder(r)
		skipped := 0
		var f mp3.Frame
		for {
			if err := d.Decode(&f, &skipped); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			duration += f.Duration()
		}
		file.Duration = duration
		// TODO: Support more formats
	}
	return nil
}

// analyzeAudio sets video-specific metadata on file.
func analyzeVideo(r io.Reader, mediaType mediatype.MediaType, file *entity.File) error {
	switch mediaType.Subtype() {
	case "mp4":
		var rs io.ReadSeeker
		var ok bool
		if rs, ok = r.(io.ReadSeeker); !ok {
			buf, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			rs = bytes.NewReader(buf)
		}
		// FIXME: Do we really have to buffer the whole file?
		info, err := mp4.Probe(rs)
		if err != nil {
			return err
		}
		file.Duration = time.Duration(info.Duration) * time.Millisecond
		for _, track := range info.Tracks {
			if track.AVC != nil {
				file.Width = int(track.AVC.Width)
				file.Height = int(track.AVC.Height)
				break
			}
		}
	}
	return nil
}

// OpenFile is passed on directly to s.store.
func (s *service) OpenFile(ctx context.Context, file *model.File) (io.ReadCloser, error) {
	return s.store.OpenFile(ctx, file.Type, file.UUID)
}
