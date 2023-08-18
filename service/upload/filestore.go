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

// FileStore is an implementation of the Store interface that saves files in the local filesystem.
// Each upload is stored in a separate folder that contains its files.
type FileStore struct {
	root string

	// These modes are applied to new files and directories.
	FileMode fs.FileMode
	DirMode  fs.FileMode
}

// NewFileStore creates a new FileStore rooted at root.
// The root directory must exist.
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

// Create opens a writer to the named file.
func (s *FileStore) Create(_ context.Context, upload uuid.UUID, name string) (io.WriteCloser, error) {
	if !fs.ValidPath(name) || name == "." {
		return nil, fs.ErrInvalid
	}
	name = filepath.Join(s.root, upload.String(), name)
	if err := os.MkdirAll(filepath.Dir(name), s.DirMode); err != nil {
		return nil, err
	}
	return os.Create(name)
}

// Stat fetches information about a named file.
func (s *FileStore) Stat(_ context.Context, upload uuid.UUID, name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, fs.ErrInvalid
	}
	path := filepath.Join(s.root, upload.String(), name)
	stat, err := os.Stat(path)
	// the root directory should always exist
	if os.IsNotExist(err) && name == "." {
		if err = os.MkdirAll(path, s.DirMode); err != nil {
			return nil, err
		}
		stat, err = os.Stat(path)
	}
	return stat, err
}

// Open opens the named file or directory.
func (s *FileStore) Open(_ context.Context, upload uuid.UUID, name string) (fs.File, error) {
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

// Delete recursively deletes the named file.
// Empty directories are not cleaned.
func (s *FileStore) Delete(_ context.Context, upload uuid.UUID, name string) error {
	if !fs.ValidPath(name) {
		return fs.ErrInvalid
	}
	name = filepath.Join(s.root, upload.String(), name)
	return os.RemoveAll(name)
}

// folderDir implements the Dir interface for FileStore.
// The implementation caches the full contents of a directory in order to work on files in alphabetical order.
type folderDir struct {
	*os.File // underlying file

	entries []fs.DirEntry // cached entries
	marker  string        // current marker
}

// Marker returns the current marker.
func (d *folderDir) Marker() string {
	return d.marker
}

// SkipTo sets the marker.
func (d *folderDir) SkipTo(marker string) error {
	if d.marker != "" && marker < d.marker {
		return errors.New("cannot skip backwards")
	}
	d.marker = marker
	return nil
}

// ReadDir reads n entries from the current marker.
// If n <= 0, all remaining entries are read and a nil error will be returned.
// If n > 0 an io.EOF error indicates that all entries have been read.
//
// A first call to ReadDir will read the entire directory contents into memory.
// All subsequent operations only operate on the in-memory data.
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
