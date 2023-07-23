package media

import (
	"context"
	"crypto/sha256"
	"github.com/Karaoke-Manager/karman/internal/model"
	"gorm.io/gorm"
	"image"
	"io"
	"sync"

	// Supported image formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type Service interface {
	StoreImageFile(ctx context.Context, mediaType string, r io.Reader) (model.File, error)
}

func NewService(db *gorm.DB, store Store) Service {
	return service{db, store}
}

type service struct {
	db    *gorm.DB
	store Store
}

func (s service) StoreImageFile(ctx context.Context, mediaType string, r io.Reader) (file model.File, err error) {
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
	h := sha256.New()
	r = io.TeeReader(r, pw)
	r = io.TeeReader(r, h)

	wg.Add(1)
	defer wg.Wait()
	// Read image data in parallel to writing
	go func() {
		defer wg.Done()

		cfg, _, err := image.DecodeConfig(pr)
		// we need to read all potentially remaining bytes as to not block the pipe
		_, _ = io.Copy(io.Discard, pr)
		if err != nil {
			// TODO: Log unknown formats and other errors
			return
		}
		file.Width = cfg.Width
		file.Height = cfg.Height
	}()

	file.Size, err = io.Copy(w, r)
	_ = pw.Close()
	if err != nil {
		// Let background job do the cleanup
		return
	}
	file.Checksum = h.Sum(nil)

	if err = s.db.WithContext(ctx).Save(&file).Error; err != nil {
		return
	}
	return
}
