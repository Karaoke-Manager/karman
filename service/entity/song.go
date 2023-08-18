package entity

import (
	"time"

	"github.com/Karaoke-Manager/karman/model"

	"codello.dev/ultrastar"
)

// Song is the entity for songs.
type Song struct {
	Entity

	// UploadID will be set if this song belongs to an upload.
	// Depending on the associations loaded Upload will also be set.
	UploadID *uint
	Upload   *Upload `gorm:"constraint:OnDelete:CASCADE"`

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
	Start           time.Duration
	End             time.Duration
	PreviewStart    time.Duration
	MedleyStartBeat ultrastar.Beat
	MedleyEndBeat   ultrastar.Beat

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

	// Music of the Song
	MusicP1 *ultrastar.Music `gorm:"type:blob;serializer:nilGob"`
	MusicP2 *ultrastar.Music `gorm:"type:blob;serializer:nilGob"`
}

// FromSong converts the specified song into a Song value containing the same metadata.
// This is usually used when updating fields of a song.
func FromSong(song *model.Song) Song {
	return Song{
		Entity:          fromModel(song.Model),
		Gap:             song.Gap,
		VideoGap:        song.VideoGap,
		Start:           song.Start,
		End:             song.End,
		PreviewStart:    song.PreviewStart,
		MedleyStartBeat: song.MedleyStartBeat,
		MedleyEndBeat:   song.MedleyEndBeat,
		Title:           song.Title,
		Artist:          song.Artist,
		Genre:           song.Genre,
		Edition:         song.Edition,
		Creator:         song.Creator,
		Language:        song.Language,
		Year:            song.Year,
		Comment:         song.Comment,
		DuetSinger1:     song.DuetSinger1,
		DuetSinger2:     song.DuetSinger2,
		Extra:           song.CustomTags,
		MusicP1:         song.MusicP1,
		MusicP2:         song.MusicP2,
	}
}

// ToModel converts s into an equivalent model.Song instance.
func (s *Song) ToModel() *model.Song {
	if s == nil {
		return nil
	}
	return &model.Song{
		Model: s.Entity.toModel(),
		Song: &ultrastar.Song{
			Gap:             s.Gap,
			VideoGap:        s.VideoGap,
			Start:           s.Start,
			End:             s.End,
			PreviewStart:    s.PreviewStart,
			MedleyStartBeat: s.MedleyStartBeat,
			MedleyEndBeat:   s.MedleyEndBeat,
			Title:           s.Title,
			Artist:          s.Artist,
			Genre:           s.Genre,
			Edition:         s.Edition,
			Creator:         s.Creator,
			Language:        s.Language,
			Year:            s.Year,
			Comment:         s.Comment,
			DuetSinger1:     s.DuetSinger1,
			DuetSinger2:     s.DuetSinger2,
			CustomTags:      s.Extra,
			MusicP1:         s.MusicP1,
			MusicP2:         s.MusicP2,
		},
		InUpload:       s.UploadID != nil,
		AudioFile:      s.AudioFile.ToModel(),
		CoverFile:      s.CoverFile.ToModel(),
		VideoFile:      s.VideoFile.ToModel(),
		BackgroundFile: s.BackgroundFile.ToModel(),
	}
}
