package entity

import (
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
)

// Upload is the database entity for uploads.
type Upload struct {
	Entity

	Open           bool
	SongsTotal     int
	SongsProcessed int
}

// ToModel converts u into an equivalent model.Upload instance.
// The errors specify the number of errors for the upload.
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

// An UploadProcessingError indicates an error that occurred during processing of an upload.
// Each error value is associated with a single upload.
type UploadProcessingError struct {
	gorm.Model

	UploadID uint
	Upload   Upload `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// File is the path of the file that caused the error.
	File string
	// Message is an error message.
	Message string
}

// ToModel converts err into an equivalent model.UploadProcessingError instance.
func (err *UploadProcessingError) ToModel() *model.UploadProcessingError {
	return &model.UploadProcessingError{
		File:    err.File,
		Message: err.Message,
	}
}
