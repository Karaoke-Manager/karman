package model

import "gorm.io/gorm"

type UploadProcessingError struct {
	gorm.Model

	UploadID uint
	File     string
	Message  string
}

type Upload struct {
	Model

	Open           bool
	SongsTotal     int
	SongsProcessed int

	ProcessingErrors []UploadProcessingError `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func NewUpload() Upload {
	return Upload{
		Open:           true,
		SongsTotal:     -1,
		SongsProcessed: -1,
	}
}
