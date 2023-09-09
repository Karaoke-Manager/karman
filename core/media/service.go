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
	"log/slog"
	"sync"
	"time"

	"github.com/abema/go-mp4" // MP4 support
	"github.com/google/uuid"
	"github.com/lmittmann/tint"
	"github.com/tcolgate/mp3" // MP3 support

	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/streamio"
)

// service is the default Service implementation.
type service struct {
	logger *slog.Logger
	repo   Repository
	store  Store
}

// NewService creates a new Service instance using the supplied db and store.
// The default implementation will store media files in the store as well as in the DB.
// For each media file there will be an entry in the DB, the actual data however lives in the store.
func NewService(logger *slog.Logger, repo Repository, store Store) Service {
	return &service{logger, repo, store}
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
		if cErr != nil {
			s.logger.ErrorContext(ctx, "Could not close media file.", tint.Err(err))
		}
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
		s.fullAnalyzeFile(ctx, pr, mediaType, &file)
	}()

	_, err = io.Copy(w, r)
	_ = pw.Close() // make sure that the goroutine stops
	wg.Wait()
	if err != nil {
		// This probably indicates that the request was cancelled or a network error occurred
		// Let background job do the cleanup
		s.logger.ErrorContext(ctx, "Could not write media file.", "uuid", file.UUID, tint.Err(err))
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
func (s *service) fullAnalyzeFile(ctx context.Context, r io.Reader, mediaType mediatype.MediaType, file *model.File) {
	s.logger.DebugContext(ctx, "Analyzing media file.", "uuid", file.UUID, "type", mediaType)
	var size int64
	r = streamio.CountBytes(r, &size)
	h := sha256.New()
	r = io.TeeReader(r, h)

	switch mediaType.Type() {
	case "image":
		s.analyzeImage(ctx, r, mediaType, file)
	case "audio":
		s.analyzeAudio(ctx, r, mediaType, file)
	case "video":
		s.analyzeVideo(ctx, r, mediaType, file)
	default:
		s.logger.WarnContext(ctx, "Unknown media file type.", "uuid", file.UUID, "type", mediaType)
	}
	// TODO: Log unknown formats and other errors

	// Read remaining bytes for accurate size and hash.
	// We also want to make sure to read r until the end.
	if _, err := io.Copy(io.Discard, r); err != nil {
		s.logger.ErrorContext(ctx, "Could not completely read media file.", "uuid", file.UUID, "type", mediaType, tint.Err(err))
		return
	}
	file.Size = size
	file.Checksum = h.Sum(nil)
}

// analyzeImage sets image-specific metadata on file.
func (s *service) analyzeImage(ctx context.Context, r io.Reader, mediaType mediatype.MediaType, file *model.File) {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		s.logger.WarnContext(ctx, "Could not decode image file.", "uuid", file.UUID, "type", mediaType, tint.Err(err))
	}
	file.Width = cfg.Width
	file.Height = cfg.Height
}

// analyzeAudio sets audio-specific metadata on file.
func (s *service) analyzeAudio(ctx context.Context, r io.Reader, mediaType mediatype.MediaType, file *model.File) {
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
				s.logger.ErrorContext(ctx, "Could not analyze MP3 audio file.", "uuid", file.UUID, "type", mediaType, tint.Err(err))
				return
			}
			duration += f.Duration()
		}
		file.Duration = duration
	// TODO: Support more formats
	default:
		s.logger.WarnContext(ctx, "Unknown audio file type.", "uuid", file.UUID, "type", mediaType)
	}
}

// analyzeAudio sets video-specific metadata on file.
func (s *service) analyzeVideo(ctx context.Context, r io.Reader, mediaType mediatype.MediaType, file *model.File) {
	switch mediaType.Subtype() {
	case "mp4":
		var rs io.ReadSeeker
		var ok bool
		if rs, ok = r.(io.ReadSeeker); !ok {
			buf, err := io.ReadAll(r)
			if err != nil {
				s.logger.ErrorContext(ctx, "Could not read MP4 video file.", "uuid", file.UUID, "type", mediaType, tint.Err(err))
				return
			}
			rs = bytes.NewReader(buf)
		}
		// FIXME: Do we really have to buffer the whole file?
		info, err := mp4.Probe(rs)
		if err != nil {
			s.logger.ErrorContext(ctx, "Could not analyze MP4 video file.", "uuid", file.UUID, "type", mediaType, tint.Err(err))
			return
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
}

// DeleteFile deletes the file with the specified UUID from the underlying store.
// If the deletion is successful the file is also deleted from the database.
func (s *service) DeleteFile(ctx context.Context, id uuid.UUID) error {
	file, err := s.repo.GetFile(ctx, id)
	if err != nil && !errors.Is(err, core.ErrNotFound) {
		return err
	}
	if _, err = s.store.Delete(ctx, file.Type, id); err != nil {
		return err
	}
	_, err = s.repo.DeleteFile(ctx, file.UUID)
	return err
}
