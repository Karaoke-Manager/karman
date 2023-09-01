package testdata

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
)

func OpenUpload(t *testing.T, db pgxutil.DB) model.Upload {
	row, err := pgxutil.InsertRowReturning(context.TODO(), db, "uploads", map[string]any{
		"open": true,
	}, "id, uuid, created_at, updated_at", pgx.RowToStructByName[creationResult])
	if err != nil {
		t.Fatalf("testdata.OpenUpload() could not insert into the database: %s", err)
	}
	return model.Upload{
		Model: model.Model{
			UUID:      row.UUID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		State:          model.UploadStateOpen,
		SongsTotal:     -1,
		SongsProcessed: -1,
		Errors:         0,
	}
}

func NOpenUploads(t *testing.T, db pgxutil.DB, n int) {
	_, err := db.CopyFrom(context.TODO(), pgx.Identifier{"uploads"}, []string{"open"}, pgx.CopyFromSlice(n, func(i int) ([]any, error) {
		return []any{true}, nil
	}))
	if err != nil {
		t.Fatalf("test.NOpenUploads() could not insert all songs: %s", err)
	}
}

func PendingUpload(t *testing.T, db pgxutil.DB) model.Upload {
	row, err := pgxutil.InsertRowReturning(context.TODO(), db, "uploads", map[string]any{
		"open": false,
	}, "id, uuid, created_at, updated_at", pgx.RowToStructByName[creationResult])
	if err != nil {
		t.Fatalf("testdata.PendingUpload() could not insert into the database: %s", err)
	}
	return model.Upload{
		Model: model.Model{
			UUID:      row.UUID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		State:          model.UploadStatePending,
		SongsTotal:     -1,
		SongsProcessed: -1,
		Errors:         0,
	}
}

func NPendingUploads(t *testing.T, db pgxutil.DB, n int) {
	_, err := db.CopyFrom(context.TODO(), pgx.Identifier{"uploads"}, []string{"open"}, pgx.CopyFromSlice(n, func(i int) ([]any, error) {
		return []any{false}, nil
	}))
	if err != nil {
		t.Fatalf("test.NOpenUploads() could not insert all songs: %s", err)
	}
}

func ProcessingUpload(t *testing.T, db pgxutil.DB) model.Upload {
	row, err := pgxutil.InsertRowReturning(context.TODO(), db, "uploads", map[string]any{
		"open":            false,
		"songs_total":     20,
		"songs_processed": 3,
	}, "id, uuid, created_at, updated_at", pgx.RowToStructByName[creationResult])
	if err != nil {
		t.Fatalf("testdata.ProcessingUpload() could not insert into the database: %s", err)
	}
	return model.Upload{
		Model: model.Model{
			UUID:      row.UUID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		State:          model.UploadStateProcessing,
		SongsTotal:     20,
		SongsProcessed: 3,
		Errors:         0,
	}
}

func DoneUpload(t *testing.T, db pgxutil.DB) model.Upload {
	u, _ := doneUpload(t, db, 20)
	return u
}

func doneUpload(t *testing.T, db pgxutil.DB, n int) (model.Upload, int) {
	row, err := pgxutil.InsertRowReturning(context.TODO(), db, "uploads", map[string]any{
		"open":            false,
		"songs_total":     n,
		"songs_processed": n,
	}, "id, uuid, created_at, updated_at", pgx.RowToStructByName[creationResult])
	if err != nil {
		t.Fatalf("testdata.DoneUpload() could not insert into the database: %s", err)
	}
	return model.Upload{
		Model: model.Model{
			UUID:      row.UUID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		State:          model.UploadStateDone,
		SongsTotal:     n,
		SongsProcessed: n,
		Errors:         0,
	}, row.ID
}

func DoneUploadWithErrors(t *testing.T, db pgxutil.DB) model.Upload {
	row, err := pgxutil.InsertRowReturning(context.TODO(), db, "uploads", map[string]any{
		"open":            false,
		"songs_total":     4,
		"songs_processed": 4,
	}, "id, uuid, created_at, updated_at", pgx.RowToStructByName[creationResult])
	if err != nil {
		t.Fatalf("testdata.DoneUploadWithErrors() could not insert into the database: %s", err)
	}
	_, err = db.CopyFrom(context.TODO(), pgx.Identifier{"upload_errors"}, []string{"upload_id", "file", "message"}, pgx.CopyFromRows([][]any{
		{row.ID, "foo.txt", "not a valid song"},
		{row.ID, "bar.txt", "invalid note"},
	}))
	if err != nil {
		t.Fatalf("testdata.DoneUploadWithErrors() could not insert errors into the database: %s", err)
	}
	return model.Upload{
		Model: model.Model{
			UUID:      row.UUID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		State:          model.UploadStateDone,
		SongsTotal:     4,
		SongsProcessed: 4,
		Errors:         2,
	}
}
