//go:build database

package upload

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/nolog"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func Test_dbRepo_CreateUpload(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)

	upload := model.Upload{}
	err := repo.CreateUpload(context.TODO(), &upload)
	if err != nil {
		t.Errorf("CreateUpload(ctx, &upload) returned an unexpected error: %s", err)
	}
	if upload.UUID == uuid.Nil {
		t.Errorf("CreateUpload(ctx, &upload) produced upload.UUID = <uuid.Nil>, expected a valud UUID")
	}
	if upload.State != model.UploadStateOpen {
		t.Errorf("CreateUpload(ctx, &upload) produced upload.State = %q, expected %q", upload.State, model.UploadStateOpen)
	}
}

func Test_dbRepo_GetUpload(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)

	t.Run("missing", func(t *testing.T) {
		id := uuid.New()
		_, err := repo.GetUpload(context.TODO(), id)
		if err == nil {
			t.Errorf("GetUpload(ctx, %q) did not return an error, expected ErrNotFound", id)
		}
		if !errors.Is(err, core.ErrNotFound) {
			t.Errorf("GetUpload(ctx, %q) returned an unexpected error: %s, expected ErrNotFound", id, err)
		}
	})

	cases := map[string]model.Upload{
		"open":       testdata.OpenUpload(t, db),
		"pending":    testdata.PendingUpload(t, db),
		"processing": testdata.ProcessingUpload(t, db),
		"done":       testdata.DoneUpload(t, db),
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			upload, err := repo.GetUpload(context.TODO(), c.UUID)
			if err != nil {
				t.Errorf("GetUpload(ctx, %q) returned an unexpected error: %s", c.UUID, err)
				return
			}
			if upload.State != c.State {
				t.Errorf("GetUpload(ctx, %q) produced upload.State = %q, expected %q", c.UUID, upload.State, c.State)
			}
			if upload.Errors != c.Errors {
				t.Errorf("GetUpload(ctx, %q) produced upload.Errors = %d, expected %d", c.UUID, upload.Errors, c.Errors)
			}
		})
	}

	uploadWithErrors := testdata.DoneUploadWithErrors(t, db)
	t.Run("errors", func(t *testing.T) {
		upload, err := repo.GetUpload(context.TODO(), uploadWithErrors.UUID)
		if err != nil {
			t.Errorf("GetUpload(ctx, %q) returned an unexpected error: %s", uploadWithErrors.UUID, err)
			return
		}
		if upload.Errors != uploadWithErrors.Errors {
			t.Errorf("GetUpload(ctx, %q) produced upload.Errors = %d, expected %d", uploadWithErrors.UUID, upload.Errors, uploadWithErrors.Errors)
		}
	})
}

func Test_dbRepo_FindUploads(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)
	testdata.NOpenUploads(t, db, 10)
	testdata.NPendingUploads(t, db, 3)

	uploads, total, err := repo.FindUploads(context.TODO(), -1, 0)
	if err != nil {
		t.Errorf("FindUploads(ctx, -1, 0) returned an unexpected error: %s", err)
		return
	}
	if total != 13 {
		t.Errorf("FindUploads(ctx, -1, 0) = _, %d, _, expected %d", total, 13)
	}
	if len(uploads) != 13 {
		t.Errorf("FindUploads(ctx, -1, 0) returned %d uploads, expected %d", len(uploads), 13)
	}
}

func Test_dbRepo_UpdateUpload(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)

	t.Run("fields", func(t *testing.T) {
		upload := testdata.OpenUpload(t, db)
		upload.State = model.UploadStateProcessing
		upload.SongsTotal = 100
		upload.SongsProcessed = 20
		oldUpdatedAt := upload.UpdatedAt
		err := repo.UpdateUpload(context.TODO(), &upload)
		if err != nil {
			t.Errorf("UpdateUpload(ctx, &upload) returned an unexpected error: %s", err)
			return
		}
		if upload.UpdatedAt == oldUpdatedAt {
			t.Errorf("UpdateUpload(ctx, &upload) did not change upload.UpdatedAt, expected change")
		}
		if upload.SongsTotal != 100 {
			t.Errorf("UpdateUpload(ctx, &upload) produced upload.SongsTotal = %d, expected %d", upload.SongsTotal, 100)
		}
	})

	t.Run("missing", func(t *testing.T) {
		upload := model.Upload{}
		upload.UUID = uuid.New()
		err := repo.UpdateUpload(context.TODO(), &upload)
		if !errors.Is(err, core.ErrNotFound) {
			t.Errorf("UpdateUpload(ctx, &upload) returned an unexpected error: %s, expected ErrNotFound", err)
		}
	})
}

func Test_dbRepo_DeleteUpload(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)
	upload := testdata.OpenUpload(t, db)

	t.Run("success", func(t *testing.T) {
		ok, err := repo.DeleteUpload(context.TODO(), upload.UUID)
		if err != nil {
			t.Errorf("DeleteUpload(ctx, %q) returned an unexpected error: %s", upload.UUID, err)
			return
		}
		if !ok {
			t.Errorf("DeleteUpload(ctx, %q) = %t, _, expected %t", upload.UUID, ok, true)
		}

		_, err = repo.GetUpload(context.TODO(), upload.UUID)
		if !errors.Is(err, core.ErrNotFound) {
			t.Errorf("GetUpload(ctx, %q) returned an upload after it was deleted", upload.UUID)
		}
	})

	t.Run("already absent", func(t *testing.T) {
		ok, err := repo.DeleteUpload(context.TODO(), uuid.New())
		if err != nil {
			t.Errorf("DeleteUpload(ctx, <missing>) returned an unexpected error: %s", err)
			return
		}
		if ok {
			t.Errorf("DeleteUpload(ctx, <missing>) = %t, _, expected %t", ok, false)
		}
	})
}

func Test_dbRepo_CreateError(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)
	upload := testdata.ProcessingUpload(t, db)

	err := repo.CreateError(context.TODO(), &upload, model.UploadProcessingError{File: "song.txt", Message: "Invalid Encoding"})
	if err != nil {
		t.Errorf("CreateError(ctx, %q, ...) returned an unexpected error: %s", upload.UUID, err)
		return
	}
	if upload.Errors != 1 {
		t.Errorf("CreateError(ctx, %q, ...) resulted in upload.Errors = %d, expected %d", upload.UUID, upload.Errors, 1)
	}
	_, n, err := repo.GetErrors(context.TODO(), upload.UUID, -1, 0)
	if err != nil {
		t.Fatalf("CreateError(ctx, ...) succeeded, but GetErrors(ctx, %q, -1, 0) failed with an unexpected error: %s", upload.UUID)
	}
	if n != 1 {
		t.Errorf("CreateError(ctx, %q, ...) resulted in %d errors, expected %d", upload.UUID, n, 1)
	}
}

func Test_dbRepo_GetErrors(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)
	upload := testdata.DoneUploadWithErrors(t, db)

	errs, total, err := repo.GetErrors(context.TODO(), upload.UUID, -1, 0)
	if err != nil {
		t.Errorf("GetErrors(ctx, %q, -1, 0) returned an unexpected error: %s", upload.UUID, err)
		return
	}
	if total != int64(upload.Errors) {
		t.Errorf("GetErrors(ctx, %q, -1, 0) = _, %d, _, expected %d", upload.UUID, total, upload.Errors)
	}
	if len(errs) != upload.Errors {
		t.Errorf("GetErrors(ctx, %q, -1, 0) returned %d errors, expected %d", upload.UUID, len(errs), upload.Errors)
	}
}

func Test_dbRepo_ClearErrors(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)
	upload := testdata.DoneUploadWithErrors(t, db)

	ok, err := repo.ClearErrors(context.TODO(), &upload)
	if err != nil {
		t.Errorf("ClearErrors(ctx, %q) returned an unexpected error: %s", upload.UUID, err)
	}
	if !ok {
		t.Errorf("ClearErrors(ctx, %q) = %t, nil, expected %t", upload.UUID, ok, true)
	}

	ok, err = repo.ClearErrors(context.TODO(), &upload)
	if err != nil {
		t.Errorf("ClearErrors(ctx, %q) [2nd time] returned an unexpected error: %s", upload.UUID, err)
	}
	if ok {
		t.Errorf("ClearErrors(ctx, %q) = %t, nil [2nd ti], expected %t", upload.UUID, ok, false)
	}
}

func Test_dbRepo_ClearSongs(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)
	upload := testdata.DoneUploadWithSongs(t, db)

	ok, err := repo.ClearSongs(context.TODO(), &upload)
	if err != nil {
		t.Errorf("ClearSongs(ctx, %q) returned an unexpected error: %s", upload.UUID, err)
	}
	if !ok {
		t.Errorf("ClearSongs(ctx, %q) = %t, nil, expected %t", upload.UUID, ok, true)
	}

	ok, err = repo.ClearSongs(context.TODO(), &upload)
	if err != nil {
		t.Errorf("ClearSongs(ctx, %q) [2nd time] returned an unexpected error: %s", upload.UUID, err)
	}
	if ok {
		t.Errorf("ClearSongs(ctx, %q) = %t, nil [2nd time], expected %t", upload.UUID, ok, false)
	}
}

func Test_dbRepo_ClearFiles(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)
	upload := testdata.DoneUploadWithFiles(t, db)

	ok, err := repo.ClearFiles(context.TODO(), &upload)
	if err != nil {
		t.Errorf("ClearFiles(ctx, %q) returned an unexpected error: %s", upload.UUID, err)
	}
	if !ok {
		t.Errorf("ClearFiles(ctx, %q) = %t, nil, expected %t", upload.UUID, ok, true)
	}

	ok, err = repo.ClearFiles(context.TODO(), &upload)
	if err != nil {
		t.Errorf("ClearFiles(ctx, %q) [2nd time] returned an unexpected error: %s", upload.UUID, err)
	}
	if ok {
		t.Errorf("ClearFiles(ctx, %q) = %t, nil [2nd time], expected %t", upload.UUID, ok, false)
	}
}
