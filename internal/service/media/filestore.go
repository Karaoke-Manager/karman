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

type FileStore struct {
	Root       string
	FileMode   os.FileMode
	FolderMode os.FileMode
}

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
		Root:       root,
		FileMode:   0660,
		FolderMode: 0770,
	}, nil
}

func (s *FileStore) CreateFile(ctx context.Context, file model.File) (io.WriteCloser, error) {
	if file.UUID == uuid.Nil {
		return nil, ErrMissingUUID
	}
	id := file.UUID.String()
	path := filepath.Join(s.Root, id[:2], id)
	if err := os.MkdirAll(filepath.Dir(path), s.FolderMode); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, s.FileMode)
}

func (s *FileStore) ReadFile(ctx context.Context, file model.File) (io.ReadCloser, error) {
	if file.UUID == uuid.Nil {
		return nil, ErrMissingUUID
	}
	id := file.UUID.String()
	path := filepath.Join(s.Root, id[:2], id)
	return os.Open(path)
}
