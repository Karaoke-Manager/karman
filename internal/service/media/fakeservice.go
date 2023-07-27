package media

import (
	"context"
	"crypto/sha256"
	"github.com/Karaoke-Manager/karman/internal/model"
	"gorm.io/gorm"
	"io"
	"strings"
)

// FakeService is a Service implementation that only uses dummy values for file contents.
// This type is intended for testing purposes only.
type FakeService struct {
	db          *gorm.DB
	Placeholder string // The dummy content for all files
}

// NewFakeService creates a new FakeService instance and returns it.
// The placeholder will be the content of all "files".
func NewFakeService(placeholder string, db *gorm.DB) Service {
	return &FakeService{db, placeholder}
}

// StoreImageFile fully reads r and returns a file with dummy values.
// file.Type will be set to mediaType.
func (f *FakeService) StoreImageFile(ctx context.Context, mediaType string, r io.Reader) (file model.File, err error) {
	file.Type = mediaType
	file.Width = 512
	file.Height = 1080
	h := sha256.New()
	var n int64
	if n, err = io.Copy(h, r); err != nil {
		return
	}
	file.Checksum = h.Sum(nil)
	file.Size = n
	if err = f.db.WithContext(ctx).Save(&file).Error; err != nil {
		return
	}
	return
}

// ReadFile returns a reader reading the static string of f.Placeholder.
func (f *FakeService) ReadFile(ctx context.Context, file model.File) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(f.Placeholder)), nil
}
