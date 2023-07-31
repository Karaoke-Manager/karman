package media

import (
	"context"
	"fmt"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/google/uuid"
	"io"
	"os"
	"path/filepath"
)

// FileStore is an implementation of the Store interface using a directory in the local filesystem.
// The FileStore assumes that it has full control over the directory tree located at its root.
//
// FileStore creates subdirectories for files based on their UUIDs.
// The subdirectories are nested by UUID prefixes to avoid performance degradation because of the number of files in a folder.
type FileStore struct {
	root       string      // absolute path to root directory of the file store.
	FileMode   os.FileMode // mode for newly created files
	FolderMode os.FileMode // mode for newly created folders
}

// NewFileStore creates a new file store rooted at root.
// The root directory must exist and be a directory.
func NewFileStore(root string) (Store, error) {
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
		root:       root,
		FileMode:   0660,
		FolderMode: 0770,
	}, nil
}

// CreateFile opens a writer for file.
// Any necessary intermediate directories are created before this method returns.
func (s *FileStore) CreateFile(ctx context.Context, file model.File) (io.WriteCloser, error) {
	if file.UUID == uuid.Nil {
		return nil, ErrMissingUUID
	}
	id := file.UUID.String()
	path := filepath.Join(s.root, id[:2], id)
	if err := os.MkdirAll(filepath.Dir(path), s.FolderMode); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, s.FileMode)
}

// ReadFile opens a reader for file.
func (s *FileStore) ReadFile(ctx context.Context, file model.File) (io.ReadCloser, error) {
	if file.UUID == uuid.Nil {
		return nil, ErrMissingUUID
	}
	id := file.UUID.String()
	path := filepath.Join(s.root, id[:2], id)
	return os.Open(path)
}