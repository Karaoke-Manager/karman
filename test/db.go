package test

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/service/entity"
)

// NewDB creates a new database for testing purposes.
// The returned DB will be backed by an in-memory SQLite database.
// The database should not be used outside the scope of t.
func NewDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:?_pragma=foreign_keys(1)"), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err, "Failed to create in-memory SQLite database.")
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	err = db.AutoMigrate(&entity.Upload{}, &entity.UploadProcessingError{}, &entity.File{}, &entity.Song{})
	require.NoError(t, err, "Failed to migrate database.")
	return db
}
