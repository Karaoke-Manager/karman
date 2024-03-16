package upload

import (
	"context"
	"io/fs"

	"github.com/google/uuid"
)

// uploadFS implements the fs.FS interface for a single upload.
// The FS is not strictly sandboxed, meaning there are no specific security measures in place
// that restrict the FS to the specified upload.
//
// An FS instance is only valid as long as its context remains valid.
type uploadFS struct {
	store Store
	ctx   context.Context
	id    uuid.UUID
}

// Open opens the named file.
// If the file is a directory, the file will implement [fs.ReadDirFile].
func (u *uploadFS) Open(name string) (fs.File, error) {
	return u.store.Open(u.ctx, u.id, name)
}

// Stat fetches information about the named file.
func (u *uploadFS) Stat(name string) (fs.FileInfo, error) {
	return u.store.Stat(u.ctx, u.id, name)
}
