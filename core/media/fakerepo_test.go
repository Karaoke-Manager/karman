package media

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

func Test_fakeRepo_CreateFile(t *testing.T) {
	t.Parallel()

	repo := NewFakeRepository()
	file := model.File{Type: mediatype.VideoMP4}
	err := repo.CreateFile(context.TODO(), &file)
	if err != nil {
		t.Errorf("CreateFile(ctx, &file) returned an unexpected error: %s", err)
		return
	}
	if file.UUID == uuid.Nil {
		t.Errorf("CreateFile(ctx, &file) did not set file.UUID, expected non-zero value")
	}
	if file.CreatedAt.IsZero() {
		t.Errorf("CreateFile(ctx, &file) did not set file.CreatedAt, expected non-zero value")
	}
	if file.UpdatedAt.IsZero() {
		t.Errorf("CreateFile(ctx, &file) did not set file.UpdatedAt, expected non-zero value")
	}
	if file.Deleted() {
		t.Errorf("CreateFile(ctx, &song) set file.DeletedAt, expected zero value")
	}
}

func Test_fakeRepo_GetFile(t *testing.T) {
	t.Parallel()

	repo := NewFakeRepository()
	id := uuid.New()
	repo.(*fakeRepo).files[id] = model.File{
		Model: model.Model{UUID: id},
		Size:  123,
	}

	t.Run("existing", func(t *testing.T) {
		file, err := repo.GetFile(context.TODO(), id)
		if err != nil {
			t.Errorf("GetFile(ctx, %q) returned an unexpected error: %s", id, err)
			return
		}
		if file.Size != 123 {
			t.Errorf("GetFile(ctx, %q) produced file.Size = %d, expected %d", id, file.Size, 123)
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

func Test_fakeRepo_UpdateFile(t *testing.T) {
	t.Parallel()

	repo := NewFakeRepository()
	expected := model.File{}
	_ = repo.CreateFile(context.TODO(), &expected)

	t.Run("found", func(t *testing.T) {
		update := model.File{
			Model: expected.Model,
		}
		update.Size = 826
		err := repo.UpdateFile(context.TODO(), &update)
		if err != nil {
			t.Errorf("UpdateFile(ctx, &update) returned an unexpected error: %s", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		update := model.File{
			Model: model.Model{UUID: uuid.New()},
		}
		err := repo.UpdateFile(context.TODO(), &update)
		if !errors.Is(err, core.ErrNotFound) {
			t.Errorf("UpdateFile(ctx, &update) returned an unexpected error: %s, expected ErrNotFound", err)
		}
	})
}

func Test_fakeRepo_DeleteFile(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	expected := model.File{
		Model: model.Model{UUID: id},
		Size:  123,
	}
	repo := NewFakeRepository()
	repo.(*fakeRepo).files[id] = expected

	ok, err := repo.DeleteFile(context.TODO(), expected.UUID)
	if err != nil {
		t.Errorf("DeleteFile(ctx, %q) returned an unexpected error: %s", expected.UUID, err)
	}
	if !ok {
		t.Errorf("DeleteFile(ctx, %q) returned ok = %t, expected %t", expected.UUID, ok, true)
	}

	// repeat delete to test idempotency
	ok, err = repo.DeleteFile(context.TODO(), expected.UUID)
	if err != nil {
		t.Errorf("DeleteFile(ctx, %q) [2nd time] returned an unexpected error: %s", expected.UUID, err)
	}
	if ok {
		t.Errorf("DeleteFile(ctx, %q) [2nd time] returned ok = %t, expected %t", expected.UUID, ok, false)
	}
}
