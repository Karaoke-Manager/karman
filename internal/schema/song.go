package schema

import (
	"errors"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/karman/internal/model"
	"net/http"
	"time"
)

type MedleyMode string

const (
	MedleyModeOff    MedleyMode = "off"
	MedleyModeAuto   MedleyMode = "auto"
	MedleyModeManual MedleyMode = "manual"
)

type AudioFile struct {
	Type     string        `json:"type"`
	Bitrate  int           `json:"bitrate"`
	Duration time.Duration `json:"duration"`
}

type VideoFile struct {
	Type     string        `json:"type"`
	Bitrate  int           `json:"bitrate"`
	Duration time.Duration `json:"duration"`
	Width    int           `json:"width"`
	Height   int           `json:"height"`
}

type ImageFile struct {
	Type   string `json:"type"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

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

type Song struct {
	SongRW
	UUID string `json:"uuid"`

	Duet       bool       `json:"duet"`
	Audio      *AudioFile `json:"audio"`
	Video      *VideoFile `json:"video"`
	Cover      *ImageFile `json:"cover"`
	Background *ImageFile `json:"background"`
}

func FromSong(m model.Song) Song {
	song := Song{
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
			Bitrate:  m.AudioFile.Bitrate,
			Duration: m.AudioFile.Duration,
		}
	}
	if m.VideoFile != nil {
		song.Video = &VideoFile{
			Type:     m.VideoFile.Type,
			Bitrate:  m.VideoFile.Bitrate,
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

func (s *SongRW) Render(http.ResponseWriter, *http.Request) error {
	if s.Medley.Mode != MedleyModeManual {
		s.Medley.MedleyStartBeat = 0
		s.Medley.MedleyEndBeat = 0
	}
	return nil
}

func (s *SongRW) Bind(*http.Request) error {
	if s.Medley.Mode == MedleyModeManual && (s.Medley.MedleyStartBeat == 0 || s.Medley.MedleyEndBeat == 0) {
		// TODO: Maybe a defined error variable?
		return errors.New("medley mode manual can only be set with an explicit medley start and end beat")
	}
	return nil
}
