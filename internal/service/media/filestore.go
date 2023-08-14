package media

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Karaoke-Manager/server/pkg/mediatype"

	"github.com/google/uuid"
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
func (s *FileStore) CreateFile(ctx context.Context, _ mediatype.MediaType, id uuid.UUID) (io.WriteCloser, error) {
	idStr := id.String()
	path := filepath.Join(s.root, idStr[:2], idStr)
	if err := os.MkdirAll(filepath.Dir(path), s.FolderMode); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, s.FileMode)
}

// OpenFile opens a reader for file.
func (s *FileStore) OpenFile(ctx context.Context, _ mediatype.MediaType, id uuid.UUID) (io.ReadCloser, error) {
	idStr := id.String()
	path := filepath.Join(s.root, idStr[:2], idStr)
	return os.Open(path)
}
