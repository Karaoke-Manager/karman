package media

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/model"
)

// fakeRepo is a Repository implementation backed by an in-memory map.
type fakeRepo struct {
	files map[uuid.UUID]model.File
}

// NewFakeRepository creates a new Repository backed by a simple in-memory storage.
func NewFakeRepository() Repository {
	return &fakeRepo{files: make(map[uuid.UUID]model.File)}
}

// CreateFile stores file and sets its UUID, CreatedAt, and UpdatedAt.
func (f fakeRepo) CreateFile(_ context.Context, file *model.File) error {
	file.UUID = uuid.New()
	file.CreatedAt = time.Now()
	file.UpdatedAt = file.CreatedAt
	f.files[file.UUID] = *file
	return nil
}

// UpdateFile updates the stored version of file.
func (f fakeRepo) UpdateFile(_ context.Context, file *model.File) error {
	if _, ok := f.files[file.UUID]; !ok {
		return core.ErrNotFound
	}
	file.UpdatedAt = time.Now()
	f.files[file.UUID] = *file
	return nil
}
