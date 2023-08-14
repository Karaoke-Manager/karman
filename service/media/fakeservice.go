package media

import (
	"context"
	"crypto/sha256"
	"io"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/service/entity"
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

// StoreFile fully reads r and returns a file with dummy values.
// file.Type will be set to mediaType.
func (f *FakeService) StoreFile(ctx context.Context, mediaType mediatype.MediaType, r io.Reader) (*model.File, error) {
	h := sha256.New()
	n, err := io.Copy(h, r)
	if err != nil {
		return nil, err
	}
	file := entity.File{
		Type: mediaType,

		Width:    512,
		Height:   1080,
		Duration: 3 * time.Minute,

		Size:     n,
		Checksum: h.Sum(nil),
	}
	if err = f.db.WithContext(ctx).Save(&file).Error; err != nil {
		return nil, err
	}
	return file.ToModel(), nil
}

// OpenFile returns a reader reading the static string of f.Placeholder.
func (f *FakeService) OpenFile(ctx context.Context, file *model.File) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(f.Placeholder)), nil
}
