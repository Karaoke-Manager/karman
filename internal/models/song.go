package models

import (
	"github.com/Karaoke-Manager/go-ultrastar"
	"time"
)

type Song struct {
	Model

	AudioFileID      *uint
	AudioFile        *File `gorm:"constraint:OnDelete:SET NULL"`
	VideoFileID      *uint
	VideoFile        *File `gorm:"constraint:OnDelete:SET NULL"`
	CoverFileID      *uint
	CoverFile        *File `gorm:"constraint:OnDelete:SET NULL"`
	BackgroundFileID *uint
	BackgroundFile   *File `gorm:"constraint:OnDelete:SET NULL"`

	Gap             time.Duration
	VideoGap        time.Duration
	NotesGap        ultrastar.Beat
	Start           time.Duration
	End             time.Duration
	PreviewStart    time.Duration
	MedleyStartBeat ultrastar.Beat
	MedleyEndBeat   ultrastar.Beat
	CalcMedley      bool

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

	MusicP1 ultrastar.Music
	MusicP2 ultrastar.Music

	// FIXME: Custom tags?
}
