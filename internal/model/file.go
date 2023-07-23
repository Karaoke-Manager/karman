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
//
// A File should be considered read-only for the most part.
// Instead of overwriting a file a new file should be created and the old one deleted.
// In practice edits to a file should be done extremely carefully to not disrupt the integrity of the File struct.
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
	Size int64
	// Checksum is the Sha256 checksum of this file, uniquely identifying its content.
	Checksum []byte `gorm:"type:varbinary"`

	// Bitrate and Durations are set only if the file's Type identifies an audio or video file.
	Bitrate  int // in bits per second
	Duration time.Duration

	// Width and Height of the image file.
	// Set only if the file's Type identifies an image file.
	Width  int // in pixels
	Height int // in pixels
}

func NewFile() File {
	return File{}
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
