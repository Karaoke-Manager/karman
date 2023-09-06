package media

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"

	"github.com/google/uuid"
)

// FileStore is an implementation of the Store interface using a directory in the local filesystem.
// The FileStore assumes that it has full control over the directory tree located at its root.
//
// FileStore creates subdirectories for files based on their UUIDs.
// The subdirectories are nested by UUID prefixes to avoid performance degradation because of the number of files in a folder.
type FileStore struct {
	root     string      // absolute path to root directory of the file store.
	FileMode fs.FileMode // mode for newly created files
	DirMode  fs.FileMode // mode for newly created folders
}

// NewFileStore creates a new file store rooted at root.
// The root directory must exist and be a directory.
func NewFileStore(root string) (*FileStore, error) {
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
		root:     root,
		FileMode: 0660,
		DirMode:  0770,
	}, nil
}

// Root returns the absolute path to the root directory for the store.
func (s *FileStore) Root() string {
	return s.root
}

// Create opens a writer for file.
// Any necessary intermediate directories are created before this method returns.
func (s *FileStore) Create(_ context.Context, _ mediatype.MediaType, id uuid.UUID) (io.WriteCloser, error) {
	idStr := id.String()
	path := filepath.Join(s.root, idStr[:2], idStr)
	if err := os.MkdirAll(filepath.Dir(path), s.DirMode); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, s.FileMode)
}

// Open opens a reader for file.
func (s *FileStore) Open(_ context.Context, _ mediatype.MediaType, id uuid.UUID) (io.ReadCloser, error) {
	idStr := id.String()
	path := filepath.Join(s.root, idStr[:2], idStr)
	return os.Open(path)
}
