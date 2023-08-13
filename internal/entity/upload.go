package entity

import (
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/internal/model"
)

type UploadProcessingError struct {
	gorm.Model

	UploadID uint
	File     string
	Message  string
}

type Upload struct {
	Entity

	Open           bool
	SongsTotal     int
	SongsProcessed int

	ProcessingErrors []UploadProcessingError `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (u *Upload) ToModel() *model.Upload {
	errors := make([]model.UploadProcessingError, len(u.ProcessingErrors))
	for i, err := range u.ProcessingErrors {
		errors[i] = model.UploadProcessingError{
			File:    err.File,
			Message: err.Message,
		}
	}
	return &model.Upload{
		Model:            u.Entity.ToModel(),
		Open:             u.Open,
		SongsTotal:       u.SongsTotal,
		SongsProcessed:   u.SongsProcessed,
		ProcessingErrors: errors,
	}
}
