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
func (r *fakeRepo) CreateFile(_ context.Context, file *model.File) error {
	file.UUID = uuid.New()
	file.CreatedAt = time.Now()
	file.UpdatedAt = file.CreatedAt
	r.files[file.UUID] = *file
	return nil
}

// GetFile fetches the file with the specified UUID.
func (r *fakeRepo) GetFile(_ context.Context, id uuid.UUID) (model.File, error) {
	file, ok := r.files[id]
	if !ok {
		return file, core.ErrNotFound
	}
	return file, nil
}

// UpdateFile updates the stored version of file.
func (r *fakeRepo) UpdateFile(_ context.Context, file *model.File) error {
	if _, ok := r.files[file.UUID]; !ok {
		return core.ErrNotFound
	}
	file.UpdatedAt = time.Now()
	r.files[file.UUID] = *file
	return nil
}

// DeleteFile deletes the file from the repo.
func (r *fakeRepo) DeleteFile(_ context.Context, id uuid.UUID) (bool, error) {
	_, ok := r.files[id]
	delete(r.files, id)
	return ok, nil
}

// FindOrphanedFiles always returns an empty list.
func (r *fakeRepo) FindOrphanedFiles(_ context.Context, _ int64) ([]model.File, error) {
	return nil, nil
}
