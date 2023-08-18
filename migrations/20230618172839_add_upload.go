package migrations

import (
	"context"
	"database/sql"
	"path"
	"runtime"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/migrations/db"
)

// This migration adds the Upload and UploadProcessingError entities.
func init() {
	type Entity struct {
		gorm.Model
		UUID uuid.UUID `gorm:"type:uuid,uniqueIndex"`
	}

	type Upload struct {
		Entity

		Open           bool
		SongsTotal     int
		SongsProcessed int
	}

	type UploadProcessingError struct {
		gorm.Model

		UploadID uint
		Upload   Upload `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

		File    string
		Message string
	}

	up := func(ctx context.Context, _ *sql.DB) error {
		return db.Get().WithContext(ctx).Migrator().CreateTable(&Upload{}, &UploadProcessingError{})
	}

	down := func(ctx context.Context, _ *sql.DB) error {
		return db.Get().WithContext(ctx).Migrator().DropTable(&UploadProcessingError{}, &Upload{})
	}

	_, filename, _, _ := runtime.Caller(0)
	_ = db.FS.WriteFile(path.Base(filename), []byte(""), 0444)
	goose.AddMigrationNoTxContext(up, down)
}
