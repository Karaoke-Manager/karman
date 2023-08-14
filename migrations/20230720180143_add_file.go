package migrations

import (
	"context"
	"database/sql"
	"path"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/migrations/db"
)

// This migration adds the File entity.
func init() {
	type Model struct {
		gorm.Model
		UUID uuid.UUID `gorm:"type:uuid,uniqueIndex"`
	}

	type Upload struct {
		Model
		// Minimal entity is enough for foreign keys.
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
		if err := db.Get().WithContext(ctx).Migrator().CreateTable(&File{}); err != nil {
			return err
		}
		return db.Get().WithContext(ctx).Migrator().CreateTable(&Upload{})
	}

	down := func(ctx context.Context, _ *sql.DB) error {
		if err := db.Get().WithContext(ctx).Migrator().DropTable(&File{}); err != nil {
			return err
		}
		return db.Get().WithContext(ctx).Migrator().DropTable(&Upload{})
	}

	_, filename, _, _ := runtime.Caller(0)
	_ = db.FS.WriteFile(path.Base(filename), []byte(""), 0444)
	goose.AddMigrationNoTxContext(up, down)
}
