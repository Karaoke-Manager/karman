package model

import (
	"fmt"
	"gorm.io/gorm"
	"mime"
	"strings"
	"time"
)

// File is a model that represents a media file of a song.
// A single File may be used by multiple songs and in different "functions" (audio/video, cover/background).
type File struct {
	Model

	// UploadID is the ID of the upload this file is associated with.
	// If UploadID is set, depending on the loaded associations Upload may be set as well.
	//
	// Path will contain the file path within the upload of this file.
	// If this File does not belong to an Upload, the Path should be ignored.
	// If this is unset, the file will be at the canonical place in the file store.
	UploadID *uint
	Upload   *Upload `gorm:"constraint:OnDelete:RESTRICT"`
	Path     string

	// Media Type of the file.
	Type string
	// Filesize in bytes.
	Size uint64
	// Checksum is a checksum of this file, uniquely identifying its content.
	Checksum []byte

	// Bitrate and Durations are set only if the file's Type identifies an audio or video file.
	Bitrate  int // in bits per second
	Duration time.Duration

	// Width and Height of the image file.
	// Set only if the file's Type identifies an image file.
	Width  int // in pixels
	Height int // in pixels
}

// BeforeSave ensures that f.Type is valid.
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
