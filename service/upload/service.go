package upload

import (
	"context"
	"io"
	"io/fs"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
)

// Service provides an interface for working with uploads.
type Service interface {
	// CreateUpload creates a new, open upload.
	CreateUpload(ctx context.Context) (*model.Upload, error)

	// GetUpload fetches the upload with the specified UUID.
	// If no such upload exists, the error will be common.ErrNotFound.
	GetUpload(ctx context.Context, id uuid.UUID) (*model.Upload, error)

	// FindUploads gives a paginated view to all uploads.
	// If limit is -1, all uploads are returned.
	// This method returns the page contents, the total number of uploads and an error (if one occurred).
	FindUploads(ctx context.Context, limit int, offset int64) ([]*model.Upload, int64, error)

	// DeleteUpload deletes the upload with the specified UUID, if it exists.
	// If no such upload exists, no error is returned.
	DeleteUpload(ctx context.Context, id uuid.UUID) error

	// CreateFile creates a file in an upload, overwriting existing files at the same path.
	// If the upload for the file does not exist, common.ErrNotFound is returned.
	CreateFile(ctx context.Context, upload *model.Upload, path string) (io.WriteCloser, error)

	// StatFile gets information about the file at path.
	// If no such file exists, the error is fs.ErrNotExist.
	StatFile(ctx context.Context, upload *model.Upload, path string) (fs.FileInfo, error)

	// OpenDir opens the directory at the specified path.
	// If no such file exists, fs.ErrNotExist is returned.
	// If the path identifies a file, the error will be non-nil.
	OpenDir(ctx context.Context, upload *model.Upload, path string) (Dir, error)

	// DeleteFile recursively deletes the file or folder at path.
	// Depending on the underlying storage, empty folders may also be removed.
	// If the file does not exist, no error is returned.
	DeleteFile(ctx context.Context, upload *model.Upload, path string) error

	// GetErrors returns a paginated list of errors belonging to the upload.
	// This method is only useful after processing has finished.
	GetErrors(ctx context.Context, upload *model.Upload, limit int, offset int64) ([]*model.UploadProcessingError, int64, error)
}

// NewService creates a new Service instance.
func NewService(db *gorm.DB, store Store) Service {
	return &service{db, store}
}

// service is the default Service implementation.
type service struct {
	db    *gorm.DB
	store Store
}
