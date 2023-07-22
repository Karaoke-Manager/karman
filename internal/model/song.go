package model

import (
	"github.com/Karaoke-Manager/go-ultrastar"
	"time"
)

// Song is the model for songs.
type Song struct {
	Model

	// UploadID will be set if this song belongs to an upload.
	// Depending on the associations loaded Upload will also be set.
	UploadID *uint
	Upload   *Upload `gorm:"constraint:OnDelete:RESTRICT"`

	// The File references of the Song are only set if the corresponding file exists.
	// The *ID fields indicate whether a corresponding file exists.
	// The *File fields may or may not be set, depending on whether they have been loaded.
	AudioFileID      *uint
	AudioFile        *File `gorm:"constraint:OnDelete:SET NULL"`
	VideoFileID      *uint
	VideoFile        *File `gorm:"constraint:OnDelete:SET NULL"`
	CoverFileID      *uint
	CoverFile        *File `gorm:"constraint:OnDelete:SET NULL"`
	BackgroundFileID *uint
	BackgroundFile   *File `gorm:"constraint:OnDelete:SET NULL"`

	// Song Metadata
	Gap             time.Duration
	VideoGap        time.Duration
	NotesGap        ultrastar.Beat
	Start           time.Duration
	End             time.Duration
	PreviewStart    time.Duration
	MedleyStartBeat ultrastar.Beat
	MedleyEndBeat   ultrastar.Beat
	CalcMedley      bool

	// Song Metadata
	Title    string
	Artist   string
	Genre    string
	Edition  string
	Creator  string
	Language string
	Year     int
	Comment  string

	DuetSinger1 string
	DuetSinger2 string
	Extra       map[string]string `gorm:"type:json;serializer:json"`

	// FIXME: Should we use foreign keys here?
	// Music of the Song
	MusicP1 *ultrastar.Music `gorm:"type:blob;serializer:nilGob"`
	MusicP2 *ultrastar.Music `gorm:"type:blob;serializer:nilGob"`
}

// NewSong creates a new blank Song instance.
func NewSong() Song {
	return Song{
		CalcMedley: true,
		Extra:      map[string]string{},
		MusicP1:    ultrastar.NewMusic(),
	}
}

// IsDuet indicates whether s is a duet.
func (s *Song) IsDuet() bool {
	return s.MusicP2 != nil
}
