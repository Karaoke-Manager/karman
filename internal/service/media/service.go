package media

import (
	"context"
	"crypto/sha256"
	"errors"
	"gorm.io/gorm"
	"image"
	"io"
	"strings"
	"sync"
	"time"
	
	"github.com/Karaoke-Manager/karman/internal/model"

	// Supported image formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	// Audio Libraries.
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
	StoreFile(ctx context.Context, mediaType string, r io.Reader) (model.File, error)

	// ReadFile creates a reader that can be used to read the file.
	// It is the caller's responsibility to close the reader when done.
	ReadFile(ctx context.Context, file model.File) (io.ReadCloser, error)
}

// NewService creates a new Service instance using the supplied db and store.
// The default implementation will store media files in the store as well as in the DB.
// For each media file there will be an entry in the DB, the actual data however lives in the store.
func NewService(db *gorm.DB, store Store) Service {
	return service{db, store}
}

// service is the default Service implementation.
type service struct {
	db    *gorm.DB
	store Store
}

// StoreFile creates a new model.File in the database and then saves the data from r into the store.
// Supported media types are analyzed on the fly.
func (s service) StoreFile(ctx context.Context, mediaType string, r io.Reader) (file model.File, err error) {
	file.Type = mediaType
	// We save the file here and at the end of the method to make sure that even
	// for half-written files an entry in the DB exists.
	// This makes finding orphaned files much easier.
	if err = s.db.WithContext(ctx).Save(&file).Error; err != nil {
		return
	}

	var w io.WriteCloser
	w, err = s.store.CreateFile(ctx, file)
	if err != nil {
		// Let background job do the cleanup
		return
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
	defer wg.Wait()
	// Analyze data in parallel to writing
	go func() {
		defer wg.Done()
		if err := fullAnalyzeFile(pr, mediaType, &file); err != nil {
			// TODO: Log unexpected error
			return
		}
	}()

	_, err = io.Copy(w, r)
	_ = pw.Close() // always make sure that the goroutine stops
	if err != nil {
		// This probably indicates that the request was cancelled or a network error occurred
		// Let background job do the cleanup
		// TODO: Log error
		return
	}

	if err = s.db.WithContext(ctx).Save(&file).Error; err != nil {
		return
	}
	return
}

// fullAnalyzeFile reads the complete data from r and updates file to reflect its contents.
// Analysis includes fields like file size and checksum but
// also performs content-specific analysis for images, video and audio files.
func fullAnalyzeFile(r io.Reader, mediaType string, file *model.File) error {
	var size int64
	r = countBytes(r, &size)
	h := sha256.New()
	r = io.TeeReader(r, h)

	if strings.HasPrefix(mediaType, "image/") {
		_ = analyzeImage(r, mediaType, file)
	} else if strings.HasPrefix(mediaType, "audio/") {
		_ = analyzeAudio(r, mediaType, file)
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
func analyzeImage(r io.Reader, mediaType string, file *model.File) error {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return err
	}
	file.Width = cfg.Width
	file.Height = cfg.Height
	return nil
}

// analyzeAudio sets audio-specific metadata on file.
func analyzeAudio(r io.Reader, mediaType string, file *model.File) error {
	switch mediaType {
	case "audio/mpeg", "audio/mpeg3", "audio/x-mpeg-3", "audio/mp3":
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

// ReadFile is passed on directly to s.store.
func (s service) ReadFile(ctx context.Context, file model.File) (io.ReadCloser, error) {
	return s.store.ReadFile(ctx, file)
}
