//go:build database

package upload

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/nolog"
	svc "github.com/Karaoke-Manager/karman/service"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func TestService_CreateUpload(t *testing.T) {
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

func TestService_GetUpload(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(nolog.Logger, db)

	t.Run("missing", func(t *testing.T) {
		id := uuid.New()
		_, err := repo.GetUpload(context.TODO(), id)
		if err == nil {
			t.Errorf("GetUpload(ctx, %q) did not return an error, expected ErrNotFound", id)
		}
		if !errors.Is(err, svc.ErrNotFound) {
			t.Errorf("GetUpload(ctx, %q) returned an unexpected error: %s", id, err)
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

func TestService_FindUploads(t *testing.T) {
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

func TestService_DeleteUpload(t *testing.T) {
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
		if !errors.Is(err, svc.ErrNotFound) {
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

func TestService_GetErrors(t *testing.T) {
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
