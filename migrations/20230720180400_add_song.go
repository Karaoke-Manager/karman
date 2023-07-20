package migrations

import (
	"context"
	"database/sql"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/karman/migrations/db"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
	"runtime"
	"time"
)

// This migration adds the Song model.
func init() {
	type Model struct {
		gorm.Model
		UUID uuid.UUID `gorm:"type:uuid,uniqueIndex"`
	}

	type File struct {
		Model
		// Minimal model is enough for foreign keys
	}

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

		MusicP1 *ultrastar.Music `gorm:"type:blob"`
		MusicP2 *ultrastar.Music `gorm:"type:blob"`
	}

	up := func(ctx context.Context, _ *sql.DB) error {
		err := db.Get().WithContext(ctx).Migrator().CreateTable(&Song{})
		return err
	}

	down := func(ctx context.Context, _ *sql.DB) error {
		err := db.Get().WithContext(ctx).Migrator().DropTable(&Song{})
		return err
	}

	_, filename, _, _ := runtime.Caller(0)
	goose.AddNamedMigrationNoTxContext(filename, up, down)
}
