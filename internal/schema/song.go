package schema

import (
	"errors"
	"net/http"
	"time"

	"codello.dev/ultrastar"
	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/internal/model"
)

// MedleyMode indicates how a song's medley is to be calculated.
type MedleyMode string

const (
	// MedleyModeOff disables medley calculation.
	// The song will not have a medley.
	MedleyModeOff MedleyMode = "off"

	// MedleyModeAuto enables automatic medley detection by UltraStar.
	// This is the default.
	MedleyModeAuto MedleyMode = "auto"

	// MedleyModeManual disables automatic medley calculation but provides manual medley start and end times.
	MedleyModeManual MedleyMode = "manual"
)

// AudioFile contains data about an audio file.
type AudioFile struct {
	Type     string        `json:"type"` // RFC 6838 media type
	Duration time.Duration `json:"duration"`
}

// VideoFile contains data about a video file.
type VideoFile struct {
	Type     string        `json:"type"` // RFC 6838 media type
	Duration time.Duration `json:"duration"`
	Width    int           `json:"width"`  // in pixels
	Height   int           `json:"height"` // in pixels
}

// ImageFile contains data about an image file.
type ImageFile struct {
	Type   string `json:"type"`   // RFC 6838 media type
	Width  int    `json:"width"`  // in pixels
	Height int    `json:"height"` // in pixels
}

// SongRW is the main schema for working with songs.
// All fields in SongRW are readable and writeable fields.
// The Song schema extends this with some read-only fields.
type SongRW struct {
	Title    string `json:"title"`
	Artist   string `json:"artist,omitempty"`
	Genre    string `json:"genre,omitempty"`
	Edition  string `json:"edition,omitempty"`
	Creator  string `json:"creator,omitempty"`
	Language string `json:"language,omitempty"`
	Year     int    `json:"year,omitempty"`
	Comment  string `json:"comment,omitempty"`

	DuetSinger1 string            `json:"duetSinger1,omitempty"`
	DuetSinger2 string            `json:"duetSinger2,omitempty"`
	Extra       map[string]string `json:"extra,omitempty"`

	Gap          time.Duration  `json:"gap,omitempty"`
	VideoGap     time.Duration  `json:"videoGap,omitempty"`
	NotesGap     ultrastar.Beat `json:"notesGap,omitempty"`
	Start        time.Duration  `json:"start,omitempty"`
	End          time.Duration  `json:"end,omitempty"`
	PreviewStart time.Duration  `json:"previewStart,omitempty"`

	Medley struct {
		Mode            MedleyMode     `json:"mode"`
		MedleyStartBeat ultrastar.Beat `json:"medleyStartBeat,omitempty"`
		MedleyEndBeat   ultrastar.Beat `json:"medleyEndBeat,omitempty"`
	} `json:"medley"`
}

// Song extends SongRW with additional read-only fields used in API responses.
// The Song schema should not be used as request schema.
type Song struct {
	SongRW
	UUID uuid.UUID `json:"uuid"`

	Duet       bool       `json:"duet"`
	Audio      *AudioFile `json:"audio"`
	Video      *VideoFile `json:"video"`
	Cover      *ImageFile `json:"cover"`
	Background *ImageFile `json:"background"`
}

// FromSong converts m into a schema instance representing the current state of m.
func FromSong(m model.Song) Song {
	song := Song{
		UUID: m.UUID,
		SongRW: SongRW{
			Title:    m.Title,
			Artist:   m.Artist,
			Genre:    m.Genre,
			Edition:  m.Edition,
			Creator:  m.Creator,
			Language: m.Language,
			Year:     m.Year,
			Comment:  m.Comment,

			DuetSinger1: m.DuetSinger1,
			DuetSinger2: m.DuetSinger2,
			Extra:       m.Extra,

			Gap:          m.Gap,
			VideoGap:     m.VideoGap,
			NotesGap:     m.NotesGap,
			Start:        m.Start,
			End:          m.End,
			PreviewStart: m.PreviewStart,
		},
		Duet: m.IsDuet(),
	}

	if !m.CalcMedley {
		song.Medley.Mode = MedleyModeOff
	} else if m.MedleyStartBeat != 0 && m.MedleyEndBeat != 0 {
		song.Medley.Mode = MedleyModeManual
		song.Medley.MedleyStartBeat = m.MedleyStartBeat
		song.Medley.MedleyEndBeat = m.MedleyEndBeat
	} else {
		song.Medley.Mode = MedleyModeAuto
	}

	if m.AudioFile != nil {
		song.Audio = &AudioFile{
			Type:     m.AudioFile.Type,
			Duration: m.AudioFile.Duration,
		}
	}
	if m.VideoFile != nil {
		song.Video = &VideoFile{
			Type:     m.VideoFile.Type,
			Duration: m.VideoFile.Duration,
			Width:    m.VideoFile.Width,
			Height:   m.VideoFile.Height,
		}
	}
	if m.CoverFile != nil {
		song.Cover = &ImageFile{
			Type:   m.CoverFile.Type,
			Width:  m.CoverFile.Width,
			Height: m.CoverFile.Height,
		}
	}
	if m.BackgroundFile != nil {
		song.Background = &ImageFile{
			Type:   m.BackgroundFile.Type,
			Width:  m.BackgroundFile.Width,
			Height: m.BackgroundFile.Height,
		}
	}
	return song
}

// Apply stores the fields of s into the respective fields of m.
func (s *SongRW) Apply(m *model.Song) {
	m.Title = s.Title
	m.Artist = s.Artist
	m.Genre = s.Genre
	m.Edition = s.Edition
	m.Creator = s.Creator
	m.Language = s.Language
	m.Year = s.Year
	m.Comment = s.Comment

	m.DuetSinger1 = s.DuetSinger1
	m.DuetSinger2 = s.DuetSinger2
	m.Extra = s.Extra

	m.Gap = s.Gap
	m.VideoGap = s.VideoGap
	m.NotesGap = s.NotesGap
	m.Start = s.Start
	m.End = s.End
	m.PreviewStart = s.PreviewStart

	switch s.Medley.Mode {
	case MedleyModeAuto:
		m.CalcMedley = true
		m.MedleyStartBeat = 0
		m.MedleyEndBeat = 0
	case MedleyModeManual:
		m.MedleyStartBeat = s.Medley.MedleyStartBeat
		m.MedleyEndBeat = s.Medley.MedleyEndBeat
	case MedleyModeOff:
		m.CalcMedley = false
	}
}

// Render implements the render.Renderer interface.
// Render makes sure that medley information is generated consistently.
func (s *SongRW) Render(http.ResponseWriter, *http.Request) error {
	if s.Medley.Mode != MedleyModeManual {
		s.Medley.MedleyStartBeat = 0
		s.Medley.MedleyEndBeat = 0
	}
	return nil
}

// Bind implements the render.Bind interface.
// Bind makes sure that the medley information is valid.
func (s *SongRW) Bind(*http.Request) error {
	if s.Medley.Mode == MedleyModeManual && (s.Medley.MedleyStartBeat == 0 || s.Medley.MedleyEndBeat == 0) {
		return errors.New("medley mode manual can only be set with an explicit medley start and end beat")
	}
	return nil
}
