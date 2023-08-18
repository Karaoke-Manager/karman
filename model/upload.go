package model

type UploadState string

const (
	UploadStateOpen       UploadState = "open"
	UploadStatePending    UploadState = "pending"
	UploadStateProcessing UploadState = "processing"
	UploadStateDone       UploadState = "done"
)

type Upload struct {
	Model

	State          UploadState // read only
	SongsTotal     int         // read only
	SongsProcessed int         // read only

	ProcessingErrors []UploadProcessingError // read only
}

type UploadProcessingError struct {
	File    string
	Message string
}
