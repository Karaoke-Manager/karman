//go:build database

package media

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/nolog"
	svc "github.com/Karaoke-Manager/karman/service"
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
		if !errors.Is(err, svc.ErrNotFound) {
			t.Errorf("UpdateFile(ctx, &file) returned an unexpected error: %s, expected ErrNotFound", err)
		}
	})

}
