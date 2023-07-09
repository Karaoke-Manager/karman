package models

import (
	"gorm.io/gorm"
	"strings"
)

type File struct {
	Model
	Size   int64
	Format string // The lower case file extension

	// Audio & Video
	Bitrate int

	// Videos & Images
	Width  int
	Height int

	// FIXME: Maybe include arbitrary metadata or EXIF data
}

func (f *File) BeforeSave(tx *gorm.DB) error {
	f.Format = strings.ToLower(f.Format)
	return nil
}
