package test

import (
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

// NewDB creates a new database for testing purposes.
// The returned DB will be backed by an in-memory SQLite database.
// The database should not be used outside the scope of t.
func NewDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to create in-memory SQLite database.")
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	err = db.AutoMigrate(&model.Song{}, &model.File{}, &model.Upload{})
	require.NoError(t, err, "Failed to migrate database.")
	return db
}
