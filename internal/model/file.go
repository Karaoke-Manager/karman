package model

import (
	"fmt"
	"gorm.io/gorm"
	"mime"
	"strings"
	"time"
)

type File struct {
	Model

	UploadID *uint
	Upload   *Upload `gorm:"constraint:OnDelete:CASCADE"`
	Path     string

	Size     uint64
	Checksum []byte
	Type     string // Mime type of the file. Must not contain parameters

	// Audio & Video
	Bitrate  int
	Duration time.Duration

	// Videos & Images
	Width  int
	Height int

	// FIXME: Maybe include arbitrary metadata or EXIF data
}

func (f *File) BeforeSave(*gorm.DB) error {
	t, _, err := mime.ParseMediaType(f.Type)
	if err != nil {
		return fmt.Errorf("file: invalid media type: %s", f.Type)
	}
	if !strings.Contains(t, "/") {
		return fmt.Errorf("file: invalid media type: %s", f.Type)
	}
	f.Type = t
	return nil
}
