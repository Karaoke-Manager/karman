package model

type UploadProcessingError struct {
	File    string
	Message string
}

type Upload struct {
	Model

	Open           bool
	SongsTotal     int
	SongsProcessed int

	ProcessingErrors []UploadProcessingError
}
