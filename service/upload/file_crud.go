package upload

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/Karaoke-Manager/karman/model"
)

// CreateFile creates a file in an upload.
func (s *service) CreateFile(ctx context.Context, upload *model.Upload, path string) (io.WriteCloser, error) {
	if path == "" || path == "." {
		return nil, fs.ErrInvalid
	}
	return s.store.Create(ctx, upload.UUID, path)
}

// StatFile returns information about the file at path.
func (s *service) StatFile(ctx context.Context, upload *model.Upload, path string) (fs.FileInfo, error) {
	return s.store.Stat(ctx, upload.UUID, path)
}

// OpenDir opens a directory for listing its contents.
func (s *service) OpenDir(ctx context.Context, upload *model.Upload, path string) (Dir, error) {
	f, err := s.store.Open(ctx, upload.UUID, path)
	if err != nil {
		return nil, err
	}
	dir, ok := f.(Dir)
	if !ok {
		return nil, fmt.Errorf("file at %s is not a directory", path)
	}
	return dir, nil
}

// DeleteFile recursively deletes a file or directory in the upload.
func (s *service) DeleteFile(ctx context.Context, upload *model.Upload, path string) error {
	if path == "" || path == "." {
		return fs.ErrInvalid
	}
	return s.store.Delete(ctx, upload.UUID, path)
}
