package model

// UploadState indicates in which processing state an upload currently is.
type UploadState string

const (
	// UploadStateOpen indicates that an upload has not been scheduled for processing yet.
	UploadStateOpen UploadState = "open"

	// UploadStatePending indicates that an upload has been marked for processing, but processing has not started yet.
	UploadStatePending UploadState = "pending"

	// UploadStateProcessing indicates that an upload is currently being processed.
	UploadStateProcessing UploadState = "processing"

	// UploadStateDone indicates that processing has finished.
	UploadStateDone UploadState = "done"
)

// Upload represents a batch upload of potentially many songs at once.
// An Upload acts like a write-only file share that a user can upload files to.
// After all files have been uploaded the upload can be marked to be processed by the Karman system.
// After processing has finished, it is possible to fetch all songs found in the upload.
type Upload struct {
	Model

	State UploadState // read only

	// The total number of songs found in an upload.
	// -1 if not yet known.
	SongsTotal int // read only

	// The number of songs processed (out of the total number of songs).
	// -1 if processing has not started yet.
	SongsProcessed int // read only

	// The number of errors that occurred during processing.
	Errors int
}

// A UploadProcessingError indicates some error that occurred during processing of an upload.
type UploadProcessingError struct {
	// The file in which the error occurred.
	File string
	// The error message.
	Message string
}

// Error returns the error message of the error.
func (err *UploadProcessingError) Error() string {
	return err.Message
}
