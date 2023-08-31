package entity

import (
	"time"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// File is a entity that represents a media file of a song.
// A single File may be used by multiple songs and in different "functions" (audio/video, cover/background).
//
// A File should be considered read-only for the most part.
// Instead of overwriting a file a new file should be created and the old one deleted.
// In practice edits to a file should be done extremely carefully to not disrupt the integrity of the File struct.
type File struct {
	Entity

	// UploadID is the ID of the upload this file is associated with.
	// If UploadID is set, depending on the loaded associations Upload may be set as well.
	//
	// Path will contain the file path within the upload of this file.
	// If this File does not belong to an Upload, the Path should be ignored.
	// If this is unset, the file will be at the canonical place in the file store.
	UploadID *uint
	Upload   *Upload `gorm:"constraint:OnDelete:CASCADE"`
	Path     string

	// Media Type of the file.
	Type mediatype.MediaType
	// Filesize in bytes.
	Size int64
	// Checksum is the Sha256 checksum of this file, uniquely identifying its content.
	Checksum []byte `gorm:"type:varbinary"`

	// Duration is set only if the file's Type identifies an audio or video file.
	Duration time.Duration

	// Width and Height of the image file.
	// Set only if the file's Type identifies an image or video file.
	Width  int // in pixels
	Height int // in pixels
}

// ToModel converts f into an equivalent model.File.
func (f *File) ToModel() *model.File {
	if f == nil {
		return nil
	}
	return &model.File{
		Model:    f.Entity.toModel(),
		Type:     f.Type,
		Size:     f.Size,
		Checksum: f.Checksum,
		Duration: f.Duration,
		Width:    f.Width,
		Height:   f.Height,
	}
}
