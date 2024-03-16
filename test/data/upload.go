//go:build database

package testdata

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// insertUpload inserts upload into the database.
// You can specify additional column values via the extra map.
func insertUpload(db pgxutil.DB, upload *model.Upload, extra map[string]any) (int, error) {
	values := map[string]any{
		"open":            upload.State == model.UploadStateOpen,
		"songs_total":     upload.SongsTotal,
		"songs_processed": upload.SongsProcessed,
	}
	for key, value := range extra {
		values[key] = value
	}
	row, err := pgxutil.InsertRowReturning(context.TODO(), db, "uploads",
		values, "id, uuid, created_at, updated_at", pgx.RowToStructByName[creationResult])
	if err != nil {
		return 0, err
	}
	upload.UUID = row.UUID
	upload.CreatedAt = row.CreatedAt
	upload.UpdatedAt = row.UpdatedAt
	return row.ID, nil
}

// OpenUpload inserts a new open upload into the database and returns it.
func OpenUpload(t *testing.T, db pgxutil.DB) model.Upload {
	upload := model.Upload{
		State:          model.UploadStateOpen,
		SongsTotal:     -1,
		SongsProcessed: -1,
	}
	if _, err := insertUpload(db, &upload, nil); err != nil {
		t.Fatalf("testdata.OpenUpload() could not insert into the database: %s", err)
	}
	return upload
}

// NOpenUploads inserts n new open uploads into the database.
func NOpenUploads(t *testing.T, db pgxutil.DB, n int) {
	_, err := db.CopyFrom(context.TODO(), pgx.Identifier{"uploads"}, []string{"open"}, pgx.CopyFromSlice(n, func(_ int) ([]any, error) {
		return []any{true}, nil
	}))
	if err != nil {
		t.Fatalf("test.NOpenUploads() could not insert all songs: %s", err)
	}
}

// PendingUpload inserts a new pending upload into the database and returns it.
func PendingUpload(t *testing.T, db pgxutil.DB) model.Upload {
	upload := model.Upload{
		State:          model.UploadStatePending,
		SongsTotal:     -1,
		SongsProcessed: -1,
	}
	if _, err := insertUpload(db, &upload, nil); err != nil {
		t.Fatalf("testdata.PendingUpload() could not insert into the database: %s", err)
	}
	return upload
}

// NPendingUploads inserts n pending uploads into the database.
func NPendingUploads(t *testing.T, db pgxutil.DB, n int) {
	_, err := db.CopyFrom(context.TODO(), pgx.Identifier{"uploads"}, []string{"open"}, pgx.CopyFromSlice(n, func(_ int) ([]any, error) {
		return []any{false}, nil
	}))
	if err != nil {
		t.Fatalf("test.NOpenUploads() could not insert all songs: %s", err)
	}
}

// ProcessingUpload inserts a new upload in the processing state into the database and returns it.
func ProcessingUpload(t *testing.T, db pgxutil.DB) model.Upload {
	upload := model.Upload{
		State:          model.UploadStateProcessing,
		SongsTotal:     20,
		SongsProcessed: 3,
	}
	if _, err := insertUpload(db, &upload, nil); err != nil {
		t.Fatalf("testdata.ProcessingUpload() could not insert into the database: %s", err)
	}
	return upload
}

// DoneUpload inserts a new upload in the done state into the database and returns it.
// The upload has no errors.
func DoneUpload(t *testing.T, db pgxutil.DB) model.Upload {
	upload := model.Upload{
		State:          model.UploadStateDone,
		SongsTotal:     20,
		SongsProcessed: 20,
	}
	if _, err := insertUpload(db, &upload, nil); err != nil {
		t.Fatalf("testdata.DoneUpload() could not insert into the database: %s", err)
	}
	return upload
}

// DoneUploadWithErrors inserts a new upload in the done state into the database and returns it.
// The upload has at least one error associated with it.
func DoneUploadWithErrors(t *testing.T, db pgxutil.DB) model.Upload {
	upload := model.Upload{
		State:          model.UploadStateDone,
		SongsTotal:     4,
		SongsProcessed: 4,
	}
	id, err := insertUpload(db, &upload, nil)
	if err != nil {
		t.Fatalf("testdata.DoneUploadWithErrors() could not insert upload into the database: %s", err)
	}
	_, err = db.CopyFrom(context.TODO(), pgx.Identifier{"upload_errors"}, []string{"upload_id", "file", "message"}, pgx.CopyFromRows([][]any{
		{id, "foo.txt", "not a valid song"},
		{id, "bar.txt", "invalid note"},
	}))
	if err != nil {
		t.Fatalf("testdata.DoneUploadWithErrors() could not insert errors into the database: %s", err)
	}
	upload.Errors = 2
	return upload
}

// DoneUploadWithSongs inserts a new upload in the done state into the database and returns it.
// The upload has at least one song associated with it.
func DoneUploadWithSongs(t *testing.T, db pgxutil.DB) model.Upload {
	upload := model.Upload{
		State:          model.UploadStateDone,
		SongsTotal:     2,
		SongsProcessed: 2,
	}
	id, err := insertUpload(db, &upload, nil)
	if err != nil {
		t.Fatalf("testdata.DoneUploadWithSongs() could not insert upload into the database: %s", err)
	}
	_, err = db.CopyFrom(context.TODO(), pgx.Identifier{"songs"}, []string{"upload_id", "title"}, pgx.CopyFromRows([][]any{
		{id, "Song 1"},
		{id, "Song 2"},
	}))
	if err != nil {
		t.Fatalf("testdata.DoneUploadWithSongs() could not insert songs into the database: %s", err)
	}
	return upload
}

func DoneUploadWithFiles(t *testing.T, db pgxutil.DB) model.Upload {
	upload := model.Upload{
		State: model.UploadStateDone,
	}
	id, err := insertUpload(db, &upload, nil)
	if err != nil {
		t.Fatalf("testdata.DoneUploadWithFiles() could not insert upload into the database: %s", err)
	}
	_, err = db.CopyFrom(context.TODO(), pgx.Identifier{"files"}, []string{"upload_id", "type", "path"}, pgx.CopyFromRows([][]any{
		{id, mediatype.AudioMPEG, "/test.mp3"},
		{id, mediatype.VideoMP4, "/test.mp4"},
		{id, mediatype.ImageJPEG, "/test.jpg"},
		{id, mediatype.ImagePNG, "/test.png"},
	}))
	if err != nil {
		t.Fatalf("testdata.DoneUploadWithFiles() could not insert files into the database: %s", err)
	}
	return upload
}
