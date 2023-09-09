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

// insertFile inserts the file into the database and returns its ID.
// You can specify additional column values via the extra map.
func insertFile(db pgxutil.DB, file *model.File, extra map[string]any) (int, error) {
	if file.Checksum == nil {
		file.Checksum = make([]byte, 0)
	}
	values := map[string]any{
		"type":     file.Type,
		"size":     file.Size,
		"checksum": file.Checksum,
		"duration": file.Duration,
		"width":    file.Width,
		"height":   file.Height,
	}
	for key, value := range extra {
		values[key] = value
	}
	row, err := pgxutil.InsertRowReturning(context.TODO(), db, "files",
		values, "id, uuid, created_at, updated_at", pgx.RowToStructByName[creationResult])
	if err != nil {
		return 0, err
	}
	file.UUID = row.UUID
	file.CreatedAt = row.CreatedAt
	file.UpdatedAt = row.UpdatedAt
	return row.ID, nil
}

// AudioFile inserts a model.File into the database that corresponds to an audio file.
// The file is only created in the database, no actual file contents are created.
func AudioFile(t *testing.T, db pgxutil.DB) model.File {
	file := model.File{
		Type:     mediatype.AudioMPEG,
		Size:     42132,
		Duration: 3 * time.Minute,
	}
	_, err := insertFile(db, &file, nil)
	if err != nil {
		t.Fatalf("testdata.AudioFile() could not insert file into the database: %s", err)
	}
	return file
}

// ImageFile inserts a model.File into the database that corresponds to an image file.
// The file is only created in the database, no actual file contents are created.
func ImageFile(t *testing.T, db pgxutil.DB) model.File {
	file := model.File{
		Type:   mediatype.ImagePNG,
		Size:   312,
		Width:  512,
		Height: 862,
	}
	_, err := insertFile(db, &file, nil)
	if err != nil {
		t.Fatalf("testdata.ImageFile() could not insert file into the database: %s", err)
	}
	return file
}

// VideoFile inserts a model.File into the database that corresponds to a video file.
// The file is only created in the database, no actual file contents are created.
func VideoFile(t *testing.T, db pgxutil.DB) model.File {
	file := model.File{
		Type:     mediatype.VideoMP4,
		Size:     312,
		Duration: 2*time.Minute + 25*time.Second,
		Width:    512,
		Height:   862,
	}
	_, err := insertFile(db, &file, nil)
	if err != nil {
		t.Fatalf("testdata.ImageFile() could not insert file into the database: %s", err)
	}
	return file
}

// FileInUpload inserts an upload and a model.File into the database.
// The file belongs to the upload.
// The file is only created in the database, no actual file contents are created.
func FileInUpload(t *testing.T, db pgxutil.DB) model.File {
	id, err := insertUpload(db, &model.Upload{State: model.UploadStateOpen}, nil)
	if err != nil {
		t.Fatalf("testdata.FileInUpload() could not insert upload into the database: %s", err)
	}
	file := model.File{
		UploadPath: "/foo/bar.mp3",
		Type:       mediatype.AudioMPEG,
		Size:       92773,
		Duration:   4 * time.Minute,
	}
	_, err = insertFile(db, &file, map[string]any{
		"upload_id": id,
		"path":      file.UploadPath,
	})
	if err != nil {
		t.Fatalf("testdata.FileInUpload() could not insert file into the database: %s", err)
	}
	return file
}
