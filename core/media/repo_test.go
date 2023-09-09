//go:build database

package media

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/nolog"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func Test_dbRepo_CreateFile(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)
	existing := testdata.AudioFile(t, db)

	t.Run("success", func(t *testing.T) {
		file := model.File{Type: mediatype.AudioMPEG}
		err := repo.CreateFile(context.TODO(), &file)
		if err != nil {
			t.Errorf("CreateFile(ctx, &file) returned an unexpected error: %s", err)
			return
		}
		if file.UUID == uuid.Nil {
			t.Errorf("CreateFile(ctx, &file) produced file.UUID = <uuid.Nil>, expected a valid UUID")
		}
		if file.CreatedAt.IsZero() {
			t.Errorf("CreateFile(ctx, &file) produced file.CreatedAt = 0, expected a valid date")
		}
		if file.UpdatedAt.IsZero() {
			t.Errorf("CreateFile(ctx, &file) produced file.UpdatedAt = 0, expected a valid date")
		}
		if !file.DeletedAt.IsZero() {
			t.Errorf("CreateFile(ctx, &file) produced file.DeletedAt = %q, expected zero value", file.DeletedAt)
		}
	})

	t.Run("existing UUID", func(t *testing.T) {
		file := model.File{Type: mediatype.ImageJPEG}
		file.UUID = existing.UUID
		err := repo.CreateFile(context.TODO(), &file)
		if err != nil {
			t.Errorf("CreateFile(ctx, &file) returned an unexpected error: %s", err)
			return
		}
		if file.UUID == existing.UUID {
			t.Errorf("CreateFile(ctx, &file) did not change file.UUID, expected change")
		}
		if file.CreatedAt == existing.CreatedAt {
			t.Errorf("CreateFile(ctx, &file) did not change file.CreatedAt, expected change")
		}
	})
}

func Test_dbRepo_GetFile(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)

	t.Run("regular file", func(t *testing.T) {
		expected := testdata.ImageFile(t, db)
		file, err := repo.GetFile(context.TODO(), expected.UUID)
		if err != nil {
			t.Errorf("GetFile(ctx, %q) returned an unexpected error: %s", expected.UUID, err)
			return
		}
		if !file.Type.Equals(expected.Type) {
			t.Errorf("GetFile(ctx, %q) produced file.Type = %q, expected %q", expected.UUID, file.Type, expected.Type)
		}
		if file.Size != expected.Size {
			t.Errorf("GetFile(ctx, %q) produced file.Size = %d, expected %d", expected.UUID, file.Size, expected.UUID)
		}
		if file.InUpload() {
			t.Errorf("GetFile(ctx, %q) produced file.InUpload() = true, expected false", expected.UUID)
		}
	})

	t.Run("upload file", func(t *testing.T) {
		expected := testdata.FileInUpload(t, db)
		file, err := repo.GetFile(context.TODO(), expected.UUID)
		if err != nil {
			t.Errorf("GetFile(ctx, %q) returned an unexpected error: %s", expected.UUID, err)
			return
		}
		if !file.Type.Equals(expected.Type) {
			t.Errorf("GetFile(ctx, %q) produced file.Type = %q, expected %q", expected.UUID, file.Type, expected.Type)
		}
		if file.Size != expected.Size {
			t.Errorf("GetFile(ctx, %q) produced file.Size = %d, expected %d", expected.UUID, file.Size, expected.UUID)
		}
		if !file.InUpload() {
			t.Errorf("GetFile(ctx, %q) produced file.InUpload() = %t, expected true", expected.UUID, file.InUpload())
		}
	})

	t.Run("missing", func(t *testing.T) {
		id := uuid.New()
		_, err := repo.GetFile(context.TODO(), id)
		if err == nil {
			t.Errorf("GetFile(ctx, %q) did not return an error, expected ErrNotFound", id)
		} else if !errors.Is(err, core.ErrNotFound) {
			t.Errorf("GetFile(ctx, %q) returned an unexpected error: %s, expected ErrNotFound", id, err)
		}
	})
}

func Test_dbRepo_UpdateFile(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)

	t.Run("fields", func(t *testing.T) {
		file := testdata.AudioFile(t, db)
		file.Size = 623
		oldUpdatedAt := file.UpdatedAt
		err := repo.UpdateFile(context.TODO(), &file)
		if err != nil {
			t.Errorf("UpdateFile(ctx, &file) returned an unexpected error: %s", err)
			return
		}
		if file.UpdatedAt == oldUpdatedAt {
			t.Errorf("UpdateFile(ctx, &file) did not change file.UpdatedAt, expected change")
		}
		if file.Size != 623 {
			t.Errorf("UpdateFile(ctx, &file) produced file.Size = %d, expected %d", file.Size, 623)
		}
	})

	t.Run("missing", func(t *testing.T) {
		file := model.File{}
		file.UUID = uuid.New()
		err := repo.UpdateFile(context.TODO(), &file)
		if !errors.Is(err, core.ErrNotFound) {
			t.Errorf("UpdateFile(ctx, &file) returned an unexpected error: %s, expected ErrNotFound", err)
		}
	})
}

func Test_dbRepo_DeleteFile(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)
	file := testdata.VideoFile(t, db)

	ok, err := repo.DeleteFile(context.TODO(), file.UUID)
	if err != nil {
		t.Errorf("DeleteFile(ctx, %q) returned an unexpected error: %s", file.UUID, err)
		return
	}
	if !ok {
		t.Errorf("DeleteFile(ctx, %q) = %t, _, expected %t", file.UUID, ok, true)
	}
	// repeat delete to test idempotency
	ok, err = repo.DeleteFile(context.TODO(), file.UUID)
	if err != nil {
		t.Errorf("DeleteFile(ctx, %q) [2nd time] returned an unexpected error: %s", file.UUID, err)
		return
	}
	if ok {
		t.Errorf("DeleteFile(ctx, %q) [2nd time] = %t, _, expected %t", file.UUID, ok, false)
	}
}

func Test_dbRepo_FindOrphanedFiles(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)
	expected := testdata.ImageFile(t, db)
	testdata.SongWithAudio(t, db)

	files, err := repo.FindOrphanedFiles(context.TODO(), -1)
	if err != nil {
		t.Errorf("FindOrphanedFiles(ctx, -1) returned an unexpected error: %s", err)
	}
	if len(files) != 1 {
		t.Errorf("FindOrphanedFiles(ctx, -1) returned %d files, expected %d", len(files), 1)
		return
	}
	if files[0].UUID != expected.UUID {
		t.Errorf("FindOrphanedFiles(ctx, -1) returned file with UUID = %s, expected %s", files[0].UUID, expected.UUID)
	}
}
