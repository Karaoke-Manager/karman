package entity

import (
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
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
	m := &model.Upload{
		Model:          u.Entity.toModel(),
		SongsTotal:     u.SongsTotal,
		SongsProcessed: u.SongsProcessed,
	}

	if u.Open {
		m.State = model.UploadStateOpen
	} else if u.SongsTotal < 0 && u.SongsProcessed < 0 {
		m.State = model.UploadStatePending
	} else if u.SongsTotal != u.SongsProcessed {
		m.State = model.UploadStateProcessing
	} else {
		m.State = model.UploadStateDone
	}

	m.ProcessingErrors = make([]model.UploadProcessingError, len(u.ProcessingErrors))
	for i, err := range u.ProcessingErrors {
		m.ProcessingErrors[i] = model.UploadProcessingError{
			File:    err.File,
			Message: err.Message,
		}
	}
	return m
}
