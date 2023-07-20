package rwfs

import (
	"io/fs"
	"os"
	"path"
)

func DirFS(root string) interface {
	FS
	MkDirFS
	RemoveFS
} {
	return dirFS{
		FS:   os.DirFS(root),
		root: root,
	}
}

type dirFS struct {
	fs.FS
	root string
}

func (dir dirFS) OpenFile(name string, flag int, perm fs.FileMode) (File, error) {
	if !fs.ValidPath(name) {
		return nil, os.ErrInvalid
	}
	p := path.Join(dir.root, name)
	f, err := os.OpenFile(p, flag, perm)
	if err != nil {
		err.(*os.PathError).Path = name
		return nil, err
	}
	return f, err
}

func (dir dirFS) MkDir(name string, perm fs.FileMode) error {
	if !fs.ValidPath(name) {
		return os.ErrInvalid
	}
	p := path.Join(dir.root, name)
	return os.Mkdir(p, perm)
}

func (dir dirFS) Remove(name string) error {
	if !fs.ValidPath(name) {
		return os.ErrInvalid
	}
	p := path.Join(dir.root, name)
	return os.Remove(p)
}

func (dir dirFS) Sub(name string) (fs.FS, error) {
	p := path.Join(dir.root, name)
	sub, err := fs.Sub(dir.FS, name)
	if err != nil {
		return nil, err
	}
	return dirFS{
		FS:   sub,
		root: p,
	}, nil
}
