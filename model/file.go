package model

import (
	"time"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// File represents a single media file that can be used by songs.
type File struct {
	Model

	// Type identifies the content type of the file, as specified by the user.
	Type mediatype.MediaType

	// File metadata is calculated automatically
	Size     int64  // read only
	Checksum []byte // read only

	Duration time.Duration // only audio and videos
	Width    int           // only images and videos
	Height   int           // only images and videos
}
