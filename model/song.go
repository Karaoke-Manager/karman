package model

import (
	"codello.dev/ultrastar"
)

// Song is the base model of Karman.
// A Song instance represents a single, singable UltraStar song.
type Song struct {
	Model

	ultrastar.Song
	Artists []string

	// InUpload indicates whether this song belongs to an upload.
	InUpload bool // read only

	// Changes to File references of a song will not be updated when the song is updated.
	// The same goes for the Song.*FileName fields.
	AudioFile      *File  // read only
	CoverFile      *File  // read only
	VideoFile      *File  // read only
	BackgroundFile *File  // read only
	TxtFileName    string // read only
}
