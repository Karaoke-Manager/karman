//go:build database

package testdata

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// AudioFile inserts a model.File into the database that corresponds to an audio file.
// The file is only created in the database, no actual file contents are created.
func AudioFile(t *testing.T, db pgxutil.DB) model.File {
	file := model.File{
		Type:     mediatype.AudioMPEG,
		Size:     42132,
		Duration: 3 * time.Minute,
	}
	row, err := pgxutil.InsertRowReturning(context.TODO(), db, "files", map[string]any{
		"type":     file.Type,
		"size":     file.Size,
		"duration": file.Duration,
	}, "id, uuid, created_at, updated_at", pgx.RowToStructByName[creationResult])
	if err != nil {
		t.Fatalf("testdata.AudioFile() could not insert into the database: %s", err)
	}
	file.UUID = row.UUID
	file.CreatedAt = row.CreatedAt
	file.UpdatedAt = row.UpdatedAt
	return file
}
