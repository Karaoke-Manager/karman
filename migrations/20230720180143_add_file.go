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

	type File struct {
		Model
		Size     uint64
		Checksum []byte
		Type     string // Mime type of the file. Must not contain parameters

		// Audio & Video
		Bitrate  int
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
