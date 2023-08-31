package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"image"
	_ "image/gif"  // GIF support
	_ "image/jpeg" // JPEG support
	_ "image/png"  // PNG support
	"io"
	"sync"
	"time"

	"github.com/abema/go-mp4" // MP4 support
	"github.com/tcolgate/mp3" // MP3 support

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/streamio"
)

// NewService creates a new Service instance using the supplied db and store.
// The default implementation will store media files in the store as well as in the DB.
// For each media file there will be an entry in the DB, the actual data however lives in the store.
func NewService(repo Repository, store Store) Service {
	return &service{repo, store}
}

// service is the default Service implementation.
type service struct {
	repo  Repository
	store Store
}

// StoreFile creates a new entity.File in the database and then saves the data from r into the store.
// Supported media types are analyzed on the fly.
func (s *service) StoreFile(ctx context.Context, mediaType mediatype.MediaType, r io.Reader) (file model.File, err error) {
	// We create the file here and update at the end of the method to make sure
	// that even for half-written files an entry in the DB exists.
	// This makes finding orphaned files much easier.
	file.Type = mediaType
	if err = s.repo.CreateFile(ctx, &file); err != nil {
		return
	}

	var w io.WriteCloser
	if w, err = s.store.Create(ctx, file.Type, file.UUID); err != nil {
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
		return
	}

	if err = s.repo.UpdateFile(ctx, &file); err != nil {
		return
	}
	return file, nil
}

// fullAnalyzeFile reads the complete data from r and updates file to reflect its contents.
// Analysis includes fields like file size and checksum but
// also performs content-specific analysis for images, video and audio files.
func fullAnalyzeFile(r io.Reader, mediaType mediatype.MediaType, file *model.File) error {
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
func analyzeImage(r io.Reader, mediaType mediatype.MediaType, file *model.File) error {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return err
	}
	file.Width = cfg.Width
	file.Height = cfg.Height
	return nil
}

// analyzeAudio sets audio-specific metadata on file.
func analyzeAudio(r io.Reader, mediaType mediatype.MediaType, file *model.File) error {
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
func analyzeVideo(r io.Reader, mediaType mediatype.MediaType, file *model.File) error {
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
