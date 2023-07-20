package model

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
	Extra       map[string]string `gorm:"type:json;serializer:json"`

	// FIXME: Maybe foreign keys for performance?
	MusicP1 *ultrastar.Music `gorm:"type:blob;serializer:nilGob"`
	MusicP2 *ultrastar.Music `gorm:"type:blob;serializer:nilGob"`
}

func (s Song) IsDuet() bool {
	return s.MusicP2 != nil
}

func NewSong() Song {
	return Song{
		CalcMedley: true,
		Extra:      map[string]string{},
		MusicP1:    ultrastar.NewMusic(),
	}
}

func NewSongWithData(data *ultrastar.Song) Song {
	return Song{
		Gap:             data.Gap,
		VideoGap:        data.VideoGap,
		NotesGap:        data.NotesGap,
		Start:           data.Start,
		End:             data.End,
		PreviewStart:    data.PreviewStart,
		MedleyStartBeat: data.MedleyStartBeat,
		MedleyEndBeat:   data.MedleyEndBeat,
		CalcMedley:      data.CalcMedley,
		Title:           data.Title,
		Artist:          data.Artist,
		Genre:           data.Genre,
		Edition:         data.Edition,
		Creator:         data.Creator,
		Language:        data.Language,
		Year:            data.Year,
		Comment:         data.Comment,
		DuetSinger1:     data.DuetSinger1,
		DuetSinger2:     data.DuetSinger2,
		Extra:           data.CustomTags,
		MusicP1:         data.MusicP1.Clone(),
		MusicP2:         data.MusicP2.Clone(),
	}
}
