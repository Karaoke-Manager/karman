package media

import (
	"context"
	"errors"
	"github.com/Karaoke-Manager/karman/internal/model"
	"io"
)

var (
	// ErrMissingUUID indicates that a file could not be saved because it does not have a UUID.
	ErrMissingUUID = errors.New("file has no UUID")
)

// Store is an interface to an underlying storage system used by Karman.
type Store interface {
	// CreateFile opens a writer for the specified file.
	// If no writer could be opened, an error will be returned.
	// It is the caller's responsibility to close the writer after writing the file contents.
	CreateFile(ctx context.Context, file model.File) (io.WriteCloser, error)

	// ReadFile opens a reader for the contents of the specified file.
	// If no reader could be opened, an error will be returned.
	// It is the caller's responsibility to close the reader after reading the file contents.
	ReadFile(ctx context.Context, file model.File) (io.ReadCloser, error)
}
