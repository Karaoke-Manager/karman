package models

type UploadStatus string

const (
	UploadStatusCreated    UploadStatus = "created"
	UploadStatusPending    UploadStatus = "pending"
	UploadStatusProcessing UploadStatus = "processing"
	UploadStatusDone       UploadStatus = "ready"
)

type Upload struct {
	Model
	Status UploadStatus
}
