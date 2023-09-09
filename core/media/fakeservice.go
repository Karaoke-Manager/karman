package media

import (
	"context"
	"crypto/sha256"
	"io"
	"time"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// fakeService is a Service implementation that only uses dummy values for file contents.
// This type is intended for testing purposes.
type fakeService struct {
	repo Repository
}

// NewFakeService creates a new fakeService instance and returns it.
// The placeholder will be the content of all "files".
func NewFakeService(repo Repository) Service {
	return &fakeService{repo}
}

// StoreFile fully reads r and returns a file with dummy values.
// file.Type will be set to mediaType.
func (f *fakeService) StoreFile(ctx context.Context, mediaType mediatype.MediaType, r io.Reader) (model.File, error) {
	h := sha256.New()
	n, err := io.Copy(h, r)
	if err != nil {
		return model.File{}, err
	}
	file := model.File{
		Type:     mediaType,
		Size:     n,
		Checksum: h.Sum(nil),
		Duration: 3 * time.Minute,
		Width:    512,
		Height:   1089,
	}
	if err = f.repo.CreateFile(ctx, &file); err != nil {
		return file, err
	}
	return file, nil
}
