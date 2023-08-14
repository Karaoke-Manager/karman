package media

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/server/pkg/mediatype"
)

// Store is an interface to an underlying storage system used by Karman.
type Store interface {
	// CreateFile opens a writer for a file with the specified media type and UUID.
	// If no writer could be opened, an error will be returned.
	// It is the caller's responsibility to close the writer after writing the file contents.
	CreateFile(ctx context.Context, mediaType mediatype.MediaType, id uuid.UUID) (io.WriteCloser, error)

	// OpenFile opens a reader for the contents of the file with the specified media type and UUID.
	// If no reader could be opened, an error will be returned.
	// It is the caller's responsibility to close the reader after reading the file contents.
	OpenFile(ctx context.Context, mediaType mediatype.MediaType, id uuid.UUID) (io.ReadCloser, error)
}
