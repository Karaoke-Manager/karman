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

// Service provides an interface for working with media files in Karman.
// An implementation of the Service interface implements the core logic associated with these files.
type Service interface {
	// StoreImageFile creates a new model.File and writes the data provided by r into the file.
	// This method should update known file metadata fields during the upload.
	//
	// If an error occurs r may have been partially consumed.
	// If any bytes have been persisted, this method must return a valid model.File that is able to identify the (potentially partial) data.
	// If the file has not been stored successfully, an error is returned.
	StoreImageFile(ctx context.Context, mediaType string, r io.Reader) (model.File, error)

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

// StoreImageFile creates a new mode.File in the database and then saves the data from r into the store.
// The image is analyzed on the fly.
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

// ReadFile is passed on directly to s.store.
func (s service) ReadFile(ctx context.Context, file model.File) (io.ReadCloser, error) {
	return s.store.ReadFile(ctx, file)
}
