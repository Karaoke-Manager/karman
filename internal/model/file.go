package model

import (
	"fmt"
	"gorm.io/gorm"
	"mime"
	"time"
)

type File struct {
	Model
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
		return fmt.Errorf("file: invalid type: %e", err)
	}
	f.Type = t
	return nil
}
