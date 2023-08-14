package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/server/internal/entity"
)

// NewDB creates a new database for testing purposes.
// The returned DB will be backed by an in-memory SQLite database.
// The database should not be used outside the scope of t.
func NewDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err, "Failed to create in-memory SQLite database.")
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	err = db.AutoMigrate(&entity.Song{}, &entity.File{}, &entity.UploadProcessingError{}, &entity.Upload{})
	require.NoError(t, err, "Failed to migrate database.")
	return db
}
