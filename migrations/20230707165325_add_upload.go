package migrations

import (
	"context"
	"database/sql"
	"github.com/Karaoke-Manager/karman/migrations/db"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

func init() {
	type Model struct {
		gorm.Model
		UUID uuid.UUID `gorm:"type:uuid,uniqueIndex"`
	}

	type Upload struct {
		Model
		Status         string
		SongsTotal     int
		SongsProcessed int
	}

	up := func(ctx context.Context, _ *sql.DB) error {
		err := db.Get().Migrator().CreateTable(&Upload{})
		return err
	}

	down := func(ctx context.Context, _ *sql.DB) error {
		err := db.Get().Migrator().DropTable(&Upload{})
		return err
	}

	goose.AddMigrationNoTxContext(up, down)
}
