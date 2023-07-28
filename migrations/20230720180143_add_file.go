package migrations

import (
	"context"
	"database/sql"
	"github.com/Karaoke-Manager/karman/migrations/db"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
	"runtime"
	"time"
)

// This migration adds the File model.
func init() {
	type Model struct {
		gorm.Model
		UUID uuid.UUID `gorm:"type:uuid,uniqueIndex"`
	}

	type Upload struct {
		Model
		// Minimal model is enough for foreign keys.
	}

	type File struct {
		Model

		UploadID *uint
		Upload   *Upload `gorm:"constraint:OnDelete:RESTRICT"`
		Path     string

		Type     string
		Size     int64
		Checksum []byte `gorm:"type:varbinary"`

		// Audio & Video
		Duration time.Duration

		// Videos & Images
		Width  int
		Height int
	}

	up := func(ctx context.Context, _ *sql.DB) error {
		err := db.Get().WithContext(ctx).Migrator().CreateTable(&File{})
		return err
	}

	down := func(ctx context.Context, _ *sql.DB) error {
		err := db.Get().WithContext(ctx).Migrator().DropTable(&File{})
		return err
	}

	_, filename, _, _ := runtime.Caller(0)
	goose.AddNamedMigrationNoTxContext(filename, up, down)
}
