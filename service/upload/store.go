package upload

import (
	"context"
	"io"
	"io/fs"

	"github.com/google/uuid"
)

type Store interface {
	Create(ctx context.Context, upload uuid.UUID, name string) (io.WriteCloser, error)
	Stat(ctx context.Context, upload uuid.UUID, name string) (fs.FileInfo, error)
	Open(ctx context.Context, upload uuid.UUID, name string) (fs.File, error)
	Delete(ctx context.Context, upload uuid.UUID, name string) error
}

type Dir interface {
	fs.ReadDirFile

	Marker() string
	SkipTo(marker string) error
}
