package upload

import (
	"context"
	"io"
	"io/fs"

	"github.com/google/uuid"
)

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
	// The create operation is not allowed for the root of an upload (".").
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
	// ReadDirFile provides the interface for reading directory contents.
	// The ReadDir method must return items in alphabetical order.
	fs.ReadDirFile

	// Marker returns the current marker.
	Marker() string

	// SkipTo sets the current marker so that a subsequent call to ReadDir will start after that marker.
	// You cannot move a marker backwards (the supplied marker must not be lexicographically less than the current Marker).
	SkipTo(marker string) error
}
