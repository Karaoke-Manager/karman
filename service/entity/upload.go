package entity

import (
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
)

type Upload struct {
	Entity

	Open           bool
	SongsTotal     int
	SongsProcessed int
}

func FromUpload(upload *model.Upload) Upload {
	u := Upload{
		Entity:         fromModel(upload.Model),
		Open:           upload.State == model.UploadStateOpen,
		SongsTotal:     upload.SongsTotal,
		SongsProcessed: upload.SongsProcessed,
	}
	return u
}

func (u *Upload) ToModel(errors int) *model.Upload {
	m := &model.Upload{
		Model:          u.Entity.toModel(),
		SongsTotal:     u.SongsTotal,
		SongsProcessed: u.SongsProcessed,
		Errors:         errors,
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
	return m
}

type UploadProcessingError struct {
	gorm.Model

	UploadID uint
	Upload   Upload `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	File     string
	Message  string
}

func (err *UploadProcessingError) ToModel() *model.UploadProcessingError {
	return &model.UploadProcessingError{
		File:    err.File,
		Message: err.Message,
	}
}
