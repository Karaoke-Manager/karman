package media

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// Service provides an interface for working with media files in Karman.
// An implementation of the Service interface implements the core logic associated with these files.
type Service interface {
	// StoreFile creates a new model.File and writes the data provided by r into the file.
	// This method updates known file metadata fields during the upload.
	// Depending on the media type implementations should analyze the file set type-specific metadata as well.
	//
	// If an error occurs r may have been partially consumed.
	// If any bytes have been persisted, this method must return a valid model.File that is able to identify the (potentially partial) data.
	// If the file has not been stored successfully, an error is returned.
	StoreFile(ctx context.Context, mediaType mediatype.MediaType, r io.Reader) (model.File, error)

	// DeleteFile removes the file with the specified UUID from the database and from the file system.
	// Implementations must make sure that a nil value is returned if and only if
	// the file was deleted from both the database and the file storage.
	DeleteFile(ctx context.Context, id uuid.UUID) error
}

// Repository is a service for storing references to model.File in a database.
type Repository interface {
	// CreateFile creates a new file reference based on the specified file.
	// The file.Type may be used to influence storage, other fields should not be used for this.
	// Implementations must set file.UUID, file.CreatedAt, and file.UpdatedAt.
	CreateFile(ctx context.Context, file *model.File) error

	// GetFile fetches the file with the specified UUID from the repository.
	GetFile(ctx context.Context, id uuid.UUID) (model.File, error)

	// UpdateFile updates the fields of file in the repository.
	UpdateFile(ctx context.Context, file *model.File) error

	// DeleteFile deletes the file with the specified UUID from the repository.
	// If no such file exists, the first return value will be false.
	//
	// This method does not delete the file contents associated with the file.
	// In most cases you should use Service.DeleteFile instead.
	DeleteFile(ctx context.Context, id uuid.UUID) (bool, error)

	// FindOrphanedFiles fetches a list of files that are not associated with a song.
	// If you pass limit < 0, all orphaned files are returned.
	FindOrphanedFiles(ctx context.Context, limit int64) ([]model.File, error)
}

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

	// Delete deletes the file with the specified media type and UUID.
	// If the file was already absent, the first return value will be false.
	Delete(ctx context.Context, mediaType mediatype.MediaType, id uuid.UUID) (bool, error)
}
