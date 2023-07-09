package models

type UploadStatus string

// TODO: Move these to the Upload Schema and possibly use relations to determine the status.
const (
	UploadStatusCreated    UploadStatus = "created"
	UploadStatusPending    UploadStatus = "pending"
	UploadStatusProcessing UploadStatus = "processing"
	UploadStatusDone       UploadStatus = "ready"
)

type Upload struct {
	Model
	Status         UploadStatus
	SongsTotal     int
	SongsProcessed int
	// TODO: Add progress reporting on the processing
	// TODO: Maybe support a quota
}
