package songs

import (
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"time"
)

type Resource struct {
	render.NopRenderer

	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Genre    string `json:"genre"`
	Edition  string `json:"edition"`
	Creator  string `json:"creator"`
	Language string `json:"language"`
	Year     int    `json:"year"`
	Comment  string `json:"comment"`

	Gap             time.Duration  `json:"gap"`
	VideoGap        time.Duration  `json:"videoGap"`
	NotesGap        ultrastar.Beat `json:"notesGap"`
	Start           time.Duration  `json:"start"`
	End             time.Duration  `json:"end"`
	PreviewStart    time.Duration  `json:"previewStart"`
	MedleyStartBeat ultrastar.Beat `json:"medleyStartBeat"`
	MedleyEndBeat   ultrastar.Beat `json:"medleyEndBeat"`
	CalcMedley      bool           `json:"calcMedley"`

	DuetSinger1 string `json:"duetSinger1"`
	DuetSinger2 string `json:"duetSinger2"`

	// Calculated fields
	Duet     bool          `json:"duet"`
	Duration time.Duration `json:"duration"`

	HasAudio      bool `json:"hasAudio"`
	HasVideo      bool `json:"hasVideo"`
	HasCover      bool `json:"hasCover"`
	HasBackground bool `json:"hasBackground"`
}
