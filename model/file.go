package model

import (
	"time"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// File represents a single media file that can be used by songs.
type File struct {
	Model

	// UploadPath indicates the path within an upload where this file exists.
	// A zero value indicates that this file does not belong to an upload.
	UploadPath string

	// Type identifies the content type of the file, as specified by the user.
	Type mediatype.MediaType

	// File metadata is calculated automatically
	Size     int64  // read only
	Checksum []byte // read only

	Duration time.Duration // only audio and videos
	Width    int           // only images and videos
	Height   int           // only images and videos
}

// InUpload indicates whether the file belongs to an upload or not.
func (f *File) InUpload() bool {
	return f.UploadPath != ""
}
