package model

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"errors"
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

	MusicP1 Music
	MusicP2 Music

	// FIXME: Custom tags?
}

type Music ultrastar.Music

func (*Music) GormDataType() string {
	return "blob"
}

func (m *Music) Scan(value any) error {
	bs, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal Music bytes")
	}

	d := gob.NewDecoder(bytes.NewReader(bs))
	if err := d.Decode(m); err != nil {
		return err
	}
	return nil
}

func (m *Music) Value() (driver.Value, error) {
	bs := &bytes.Buffer{}
	e := gob.NewEncoder(bs)
	if err := e.Encode(m); err != nil {
		return nil, err
	}
	return bs.Bytes(), nil
}
