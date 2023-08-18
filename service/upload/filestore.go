package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type FileStore struct {
	root     string
	FileMode fs.FileMode
	DirMode  fs.FileMode
}

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

func (s *FileStore) Create(ctx context.Context, upload uuid.UUID, name string) (io.WriteCloser, error) {
	if !fs.ValidPath(name) || name == "." {
		return nil, fs.ErrInvalid
	}
	name = filepath.Join(s.root, upload.String(), name)
	if err := os.MkdirAll(filepath.Dir(name), s.DirMode); err != nil {
		return nil, err
	}
	return os.Create(name)
}

func (s *FileStore) Stat(ctx context.Context, upload uuid.UUID, name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, fs.ErrInvalid
	}
	path := filepath.Join(s.root, upload.String(), name)
	stat, err := os.Stat(path)
	if os.IsNotExist(err) && name == "." {
		if err = os.MkdirAll(path, s.DirMode); err != nil {
			return nil, err
		}
		println("Retry")
		stat, err = os.Stat(path)
	}
	return stat, err
}

func (s *FileStore) Open(ctx context.Context, upload uuid.UUID, name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, fs.ErrInvalid
	}
	name = filepath.Join(s.root, upload.String(), name)
	f, err := os.Open(name)
	if err != nil {
		return f, err
	}
	stat, err := f.Stat()
	if err != nil {
		return f, err
	}
	if stat.IsDir() {
		return &folderDir{File: f}, nil
	}
	return f, nil
}

func (s *FileStore) Delete(ctx context.Context, upload uuid.UUID, name string) error {
	if !fs.ValidPath(name) {
		return fs.ErrInvalid
	}
	name = filepath.Join(s.root, upload.String(), name)
	return os.RemoveAll(name)
}

type folderDir struct {
	*os.File

	entries []fs.DirEntry
	marker  string
}

func (d *folderDir) Marker() string {
	return d.marker
}

func (d *folderDir) SkipTo(marker string) error {
	if d.marker != "" && marker < d.marker {
		return errors.New("cannot skip backwards")
	}
	d.marker = marker
	return nil
}

func (d *folderDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.entries == nil {
		entries, err := os.ReadDir(d.File.Name())
		if err != nil {
			return nil, err
		}
		d.entries = entries
	}

	start := 0
	if d.marker != "" {
		start = len(d.entries)
		for i, entry := range d.entries {
			if entry.Name() > d.marker {
				start = i
				break
			}
		}
	}
	end := min(start+n, len(d.entries))
	var err error
	if n <= 0 {
		end = len(d.entries)
	}
	if end > start {
		d.marker = d.entries[end-1].Name()
	} else if n > 0 {
		err = io.EOF
	}
	return d.entries[start:end], err
}
