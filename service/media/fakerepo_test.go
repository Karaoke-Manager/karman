package media

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	svc "github.com/Karaoke-Manager/karman/service"
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
		if !errors.Is(err, svc.ErrNotFound) {
			t.Errorf("UpdateFile(ctx, &update) returned an unexpected error: %s, expected ErrNotFound", err)
		}
	})
}
