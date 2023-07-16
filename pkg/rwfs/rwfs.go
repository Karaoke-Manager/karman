package rwfs

import (
	"io/fs"
	"os"
)

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

func Create(fsys FS, name string, perm fs.FileMode) (File, error) {
	return fsys.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
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
