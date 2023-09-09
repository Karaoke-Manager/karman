package upload

import (
	"context"
	"io"
	"io/fs"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
)

// Repository provides methods for storing uploads.
type Repository interface {
	// CreateUpload creates a new, open upload.
	// The upload is passed as parameter for consistency with other repositories.
	// However, currently no data from the upload is being used during creation.
	CreateUpload(ctx context.Context, upload *model.Upload) error

	// GetUpload fetches the upload with the specified UUID.
	// If no such upload exists, the error will be common.ErrNotFound.
	GetUpload(ctx context.Context, id uuid.UUID) (model.Upload, error)

	// FindUploads gives a paginated view to all uploads.
	// If limit is -1, all uploads are returned.
	// This method returns the page contents, the total number of uploads and an error (if one occurred).
	FindUploads(ctx context.Context, limit int, offset int64) ([]model.Upload, int64, error)

	// DeleteUpload deletes the upload with the specified UUID, if it exists.
	// If no such upload exists, the first return value will be false.
	DeleteUpload(ctx context.Context, id uuid.UUID) (bool, error)

	// GetErrors returns a paginated list of errors belonging to the upload.
	// This method is only useful after processing has finished.
	GetErrors(ctx context.Context, id uuid.UUID, limit int, offset int64) ([]model.UploadProcessingError, int64, error)
}

// Store is an interface used by the upload service to facilitate the actual file storage.
// The interface is inspired by fs.FS but unfortunately not compatible.
//
// The Store methods all take a context (for potential cancellation) and an upload UUID.
// The UUID serves as a namespace for filenames, that is the same filename can exist for different UUIDs.
// Whether different UUIDs are stored as folders or in a different manner is dependent on the implementation.
type Store interface {
	// Create creates a new file for upload.
	// If a file with the specified name already exists, it is overwritten.
	// If the returned error is nil, the writer must be closed when done.
	//
	// Calling Create for the root of an upload with name = "." is invalid.
	Create(ctx context.Context, upload uuid.UUID, name string) (io.WriteCloser, error)

	// Stat returns information about the named file or directory.
	// If no such file or directory exists, the returned error will be fs.ErrNotExist.
	Stat(ctx context.Context, upload uuid.UUID, name string) (fs.FileInfo, error)

	// Open opens the named file for reading.
	// If name designates a directory the returned file must implement the Dir interface.
	// If the file does not exist, the returned error will be fs.ErrNotExist.
	//
	// You can open the file "." to list the root directory of an upload.
	Open(ctx context.Context, upload uuid.UUID, name string) (fs.File, error)

	// Delete deletes the named file or directory.
	// Directories are deleted recursively.
	// If the named file does not exist, nil (no error) is returned.
	// If an error occurs the deletion process may or may not continue for subsequent files.
	//
	// If the last file in a folder is deleted this method may or may not delete empty directories as well.
	//
	// If name is ".", all files for the upload are deleted.
	Delete(ctx context.Context, upload uuid.UUID, name string) error
}

// Dir represents a directory in a Store.
// If Store.Open is called for a directory, the returned file must implement this interface.
//
// This type defines an interface for marker-based pagination.
// A marker is a string (typically the name of a file) indicating where the next read operation should begin.
// The marker itself is always excluded.
// This is inspired by the Amazon S3 API.
//
// The ReadDir method must return items in alphabetical order.
//
// As an example consider a directory with 4 files: dir1, dir2, file1, file2.
// If you supply a marker of "dir2", the next entry will be file1.
// If you supply a marker of "echo", the next entry will also be file1 (markers do not need to be valid files).
//
// After a ReadDir call has returned, you can inspect the current marker using the Marker method.
type Dir interface {
	fs.File

	// Readdir provides the interface for reading directory contents.
	// In general this method works like os.Readdir, with the following exceptions:
	//   - The ReadDir method must return items in alphabetical order.
	//   - ReadDir is influenced by SkipTo.
	Readdir(n int) ([]fs.FileInfo, error)

	// Marker returns the current marker.
	Marker() string

	// SkipTo sets the current marker so that a subsequent call to ReadDir will start after that marker.
	// You cannot move a marker backwards (the supplied marker must not be lexicographically less than the current Marker).
	SkipTo(marker string) error
}
