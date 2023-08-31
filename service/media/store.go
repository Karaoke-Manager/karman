package media

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// Store is an interface to an underlying storage system used by Karman.
type Store interface {
	// Create opens a writer for a file with the specified media type and UUID.
	// If no writer could be opened, an error will be returned.
	// It is the caller's responsibility to close the writer after writing the file contents.
	Create(ctx context.Context, mediaType mediatype.MediaType, id uuid.UUID) (io.WriteCloser, error)

	// Open opens a reader for the contents of the file with the specified media type and UUID.
	// If no reader could be opened, an error will be returned.
	// It is the caller's responsibility to close the reader after reading the file contents.
	Open(ctx context.Context, mediaType mediatype.MediaType, id uuid.UUID) (io.ReadCloser, error)
}
