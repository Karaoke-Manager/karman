package media

import (
	"context"
	"errors"
	"github.com/Karaoke-Manager/karman/internal/model"
	"io"
)

var (
	ErrMissingUUID = errors.New("file has no UUID")
)

type Store interface {
	CreateFile(ctx context.Context, file model.File) (io.WriteCloser, error)
}
