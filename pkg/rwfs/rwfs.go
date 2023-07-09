package rwfs

import (
	"errors"
	"io/fs"
	"os"
)

var ErrUnsupported = errors.New("rwfs: unsupported operation")

func OpenFile(fsys fs.FS, name string, flag int, perm os.FileMode) (fs.File, error) {
	if fsys, ok := fsys.(FS); ok {
		return fsys.OpenFile(name, flag, perm)
	}

	if flag == os.O_RDONLY {
		return fsys.Open(name)
	}
	return nil, ErrUnsupported
}

type FS interface {
	fs.FS
	OpenFile(name string, flag int, perm fs.FileMode) (File, error)
}

type File interface {
	fs.File
	Write(b []byte) (int, error)
}

type MkDirFS interface {
	fs.FS
	MkDir(name string, perm fs.FileMode) error
}

type RemoveFS interface {
	FS
	Remove(name string) error
}

func Create(fsys FS, name string) (File, error) {
	return fsys.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

func WriteFile(fsys FS, name string, data []byte, perm fs.FileMode) error {
	f, err := fsys.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

func MkDirAll(fsys MkDirFS, path string, perm fs.FileMode) error {
	// Use fsys.MkDir to do the work.
	// Also requires either Stat or Open to check for parents.
	// I'm not sure how to structure that either/or requirement.
}
