package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lmittmann/tint"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"

	"github.com/google/uuid"
)

// FileStore is an implementation of the Store interface using a directory in the local filesystem.
// The FileStore assumes that it has full control over the directory tree located at its root.
//
// FileStore creates subdirectories for files based on their UUIDs.
// The subdirectories are nested by UUID prefixes to avoid performance degradation because of the number of files in a folder.
type FileStore struct {
	logger   *slog.Logger
	root     string      // absolute path to root directory of the file store.
	FileMode fs.FileMode // mode for newly created files
	DirMode  fs.FileMode // mode for newly created folders
}

// NewFileStore creates a new file store rooted at root.
// The root directory must exist and be a directory.
func NewFileStore(logger *slog.Logger, root string) (*FileStore, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}
	return &FileStore{
		logger:   logger,
		root:     root,
		FileMode: 0660,
		DirMode:  0770,
	}, nil
}

// Root returns the absolute path to the root directory for the store.
func (s *FileStore) Root() string {
	return s.root
}

// filePath returns the path to the file with the specified UUID.
func (s *FileStore) filePath(id uuid.UUID) string {
	idStr := id.String()
	return filepath.Join(s.root, idStr[:2], idStr)
}

// Create opens a writer for file.
// Any necessary intermediate directories are created before this method returns.
func (s *FileStore) Create(ctx context.Context, _ mediatype.MediaType, id uuid.UUID) (io.WriteCloser, error) {
	path := s.filePath(id)
	if err := os.MkdirAll(filepath.Dir(path), s.DirMode); err != nil {
		s.logger.ErrorContext(ctx, "Could not create media file.", "uuid", id, tint.Err(err))
		return nil, err
	}
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, s.FileMode)
}

// Open opens a reader for file.
func (s *FileStore) Open(ctx context.Context, _ mediatype.MediaType, id uuid.UUID) (io.ReadCloser, error) {
	path := s.filePath(id)
	r, err := os.Open(path)
	if err != nil {
		s.logger.ErrorContext(ctx, "Could not open media file.", "uuid", id, tint.Err(err))
		return r, err
	}
	return r, nil
}

// Delete deletes the file.
func (s *FileStore) Delete(ctx context.Context, _ mediatype.MediaType, id uuid.UUID) (bool, error) {
	path := s.filePath(id)
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "Could not delete media file.", "uuid", id, tint.Err(err))
		return false, err
	}
	return true, nil
}
