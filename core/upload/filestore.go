package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"
	"github.com/lmittmann/tint"
)

// FileStore is an implementation of the Store interface that saves files in the local filesystem.
// Each upload is stored in a separate folder that contains its files.
type FileStore struct {
	logger *slog.Logger
	root   string

	// These modes are applied to new files and directories.
	FileMode fs.FileMode
	DirMode  fs.FileMode
}

// NewFileStore creates a new FileStore rooted at root.
// The root directory must exist.
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
		FileMode: 0640,
		DirMode:  0750,
	}, nil
}

// Root returns the absolute path to the root directory of the store.
func (s *FileStore) Root() string {
	return s.root
}

// Create opens a writer to the named file.
func (s *FileStore) Create(ctx context.Context, upload uuid.UUID, name string) (io.WriteCloser, error) {
	if !fs.ValidPath(name) || name == "." {
		s.logger.WarnContext(ctx, "Could not create upload file at invalid path.", "uuid", upload, "path", name)
		return nil, fs.ErrInvalid
	}
	name = filepath.Join(s.root, upload.String(), name)
	if err := os.MkdirAll(filepath.Dir(name), s.DirMode); err != nil {
		s.logger.ErrorContext(ctx, "Could not create intermediate directories for upload file.", "uuid", upload, "path", name, tint.Err(err))
		return nil, err
	}
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, s.FileMode)
	if err != nil {
		s.logger.ErrorContext(ctx, "Could not open upload file for writing.", "uuid", upload, "path", name, tint.Err(err))
		return f, err
	}
	return f, nil
}

// Stat fetches information about a named file.
func (s *FileStore) Stat(ctx context.Context, upload uuid.UUID, name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		s.logger.WarnContext(ctx, "Could not stat upload file at invalid path.", "uuid", upload, "path", name)
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
	if err != nil {
		s.logger.ErrorContext(ctx, "Could not stat upload file.", "uuid", upload, "path", name, tint.Err(err))
	}
	return stat, err
}

// Open opens the named file or directory.
func (s *FileStore) Open(ctx context.Context, upload uuid.UUID, name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		s.logger.WarnContext(ctx, "Could not open upload file at invalid path.", "uuid", upload, "path", name)
		return nil, fs.ErrInvalid
	}
	name = filepath.Join(s.root, upload.String(), name)
	f, err := os.Open(name)
	if err != nil {
		s.logger.ErrorContext(ctx, "Could not open upload file.", "uuid", upload, "path", name, tint.Err(err))
		return f, err
	}
	stat, err := f.Stat()
	if err != nil {
		s.logger.ErrorContext(ctx, "Could not stat opened upload file.", "uuid", upload, "path", name, tint.Err(err))
		return f, err
	}
	if stat.IsDir() {
		return &folderDir{File: f}, nil
	}
	return f, nil
}

// Delete recursively deletes the named file.
// Empty directories are not cleaned.
func (s *FileStore) Delete(ctx context.Context, upload uuid.UUID, name string) error {
	if !fs.ValidPath(name) {
		s.logger.WarnContext(ctx, "Could not delete upload file at invalid path.", "uuid", upload, "path", name)
		return fs.ErrInvalid
	}
	name = filepath.Join(s.root, upload.String(), name)
	if err := os.RemoveAll(name); err != nil {
		s.logger.ErrorContext(ctx, "Could not delete upload file.", "uuid", upload, "path", name, tint.Err(err))
		return err
	}
	return nil
}

// FS returns a fs.FS instance for the specified upload.
// The returned instance is bound to ctx and should not be used after ctx is invalidated or canceled.
func (s *FileStore) FS(ctx context.Context, upload uuid.UUID) fs.FS {
	return &uploadFS{s, ctx, upload}
}

// folderDir implements the Dir interface for FileStore.
// The implementation caches the full contents of a directory in order to work on files in alphabetical order.
type folderDir struct {
	*os.File // underlying file

	entries []fs.DirEntry
	infos   []fs.FileInfo // cached infos
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

// ReadDir reads n fs.DirEntry values from the current marker.
// If n <= 0, all remaining infos are read and a nil error will be returned.
// If n > 0 an io.EOF error indicates that all infos have been read.
//
// A first call to ReadDir will read the entire directory contents into memory.
// All subsequent operations only operate on the in-memory data.
func (d *folderDir) ReadDir(n int) ([]fs.DirEntry, error) {
	// TODO: Test for this method
	if d.entries == nil {
		entries, err := d.File.ReadDir(0)
		if err != nil {
			return nil, err
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
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

// Readdir reads n infos from the current marker.
// If n <= 0, all remaining infos are read and a nil error will be returned.
// If n > 0 an io.EOF error indicates that all infos have been read.
//
// A first call to Readdir will read the entire directory contents into memory.
// All subsequent operations only operate on the in-memory data.
func (d *folderDir) Readdir(n int) ([]fs.FileInfo, error) {
	if d.infos == nil {
		entries, err := d.File.Readdir(0)
		if err != nil {
			return nil, err
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		d.infos = entries
	}

	start := 0
	if d.marker != "" {
		start = len(d.infos)
		for i, entry := range d.infos {
			if entry.Name() > d.marker {
				start = i
				break
			}
		}
	}
	end := min(start+n, len(d.infos))
	var err error
	if n <= 0 {
		end = len(d.infos)
	}
	if end > start {
		d.marker = d.infos[end-1].Name()
	} else if n > 0 {
		err = io.EOF
	}
	return d.infos[start:end], err
}
