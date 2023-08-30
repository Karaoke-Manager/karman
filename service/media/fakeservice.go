package media

import (
	"context"
	"crypto/sha256"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// FakeService is a Service implementation that only uses dummy values for file contents.
// This type is intended for testing purposes only.
type FakeService struct {
	Placeholder string                    // The dummy content for all files
	files       map[uuid.UUID]*model.File // stored files
}

// NewFakeService creates a new FakeService instance and returns it.
// The placeholder will be the content of all "files".
func NewFakeService(placeholder string) Service {
	return &FakeService{placeholder, make(map[uuid.UUID]*model.File)}
}

// StoreFile fully reads r and returns a file with dummy values.
// file.Type will be set to mediaType.
func (f *FakeService) StoreFile(ctx context.Context, mediaType mediatype.MediaType, r io.Reader) (*model.File, error) {
	h := sha256.New()
	n, err := io.Copy(h, r)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	file := &model.File{
		Model: model.Model{
			UUID:      uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Type:     mediaType,
		Size:     n,
		Checksum: h.Sum(nil),
		Duration: 3 * time.Minute,
		Width:    512,
		Height:   1089,
	}
	f.files[file.UUID] = file
	return file, nil
}

// OpenFile returns a reader reading the static string of f.Placeholder.
func (f *FakeService) OpenFile(ctx context.Context, file *model.File) (io.ReadCloser, error) {
	if _, ok := f.files[file.UUID]; !ok {
		return nil, fs.ErrNotExist
	}
	return io.NopCloser(strings.NewReader(f.Placeholder)), nil
}
