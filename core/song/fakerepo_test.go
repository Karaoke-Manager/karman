package song

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/model"
)

func Test_fakeRepo_CreateSong(t *testing.T) {
	t.Parallel()

	repo := NewFakeRepository()
	song := model.Song{}
	err := repo.CreateSong(context.TODO(), &song)
	if err != nil {
		t.Errorf("CreateSong(ctx, &song) returned an unexpected error: %s", err)
		return
	}
	if song.UUID == uuid.Nil {
		t.Errorf("CreateSong(ctx, &song) did not set song.UUID, expected non-zero value")
	}
	if song.CreatedAt.IsZero() {
		t.Errorf("CreateSong(ctx, &song) did not set song.CreatedAt, expected non-zero value")
	}
	if song.UpdatedAt.IsZero() {
		t.Errorf("CreateSong(ctx, &song) did not set song.UpdatedAt, expected non-zero value")
	}
	if song.Deleted() {
		t.Errorf("CreateSong(ctx, &song) set song.DeletedAt, expected zero value")
	}
}

func Test_fakeRepo_GetSong(t *testing.T) {
	t.Parallel()

	repo := NewFakeRepository()
	expected := model.Song{}
	_ = repo.CreateSong(context.TODO(), &expected)

	t.Run("existing", func(t *testing.T) {
		song, err := repo.GetSong(context.TODO(), expected.UUID)
		if err != nil {
			t.Errorf("GetSong(ctx, %q) returned an unexpected error: %s", expected.UUID, err)
			return
		}
		if song.UUID == uuid.Nil {
			t.Errorf("GetSong(ctx, %q) returned an empty song, expected song.UUID = %q", expected.UUID, expected.UUID)
		}
	})
	t.Run("missing", func(t *testing.T) {
		id := uuid.New()
		_, err := repo.GetSong(context.TODO(), id)
		if !errors.Is(err, core.ErrNotFound) {
			t.Errorf("GetSong(ctx, %q) produced an unexpected error: %s, expected ErrNotFound", id, err)
		}
	})
}

func Test_fakeRepo_FindSongs(t *testing.T) {
	t.Parallel()

	repo := NewFakeRepository()
	for i := 0; i < 5; i++ {
		song := &model.Song{}
		_ = repo.CreateSong(context.TODO(), song)
	}

	cases := map[string]struct {
		Limit  int
		Offset int64
		Len    int
	}{
		"all":             {-1, 0, 5},
		"offset":          {-1, 2, 3},
		"offset past end": {-1, 6, 0},
		"limit zero":      {0, 0, 0},
		"limit":           {2, 0, 2},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			songs, total, err := repo.FindSongs(context.TODO(), c.Limit, c.Offset)
			if err != nil {
				t.Errorf("FindSongs(ctx, %d, %d) returned an unexpected error: %s", c.Limit, c.Offset, err)
				return
			}
			if total != 5 {
				t.Errorf("FindSongs(ctx, %d, %d) returned total = %d, expected %d", c.Limit, c.Offset, total, 5)
			}
			if len(songs) != c.Len {
				t.Errorf("FindSongs(ctx, %d, %d) returned %d songs, expected %d", c.Limit, c.Offset, len(songs), c.Len)
			}
		})
	}
}

func Test_fakeRepo_DeleteSong(t *testing.T) {
	t.Parallel()

	repo := NewFakeRepository()
	expected := model.Song{}
	_ = repo.CreateSong(context.TODO(), &expected)

	ok, err := repo.DeleteSong(context.TODO(), expected.UUID)
	if err != nil {
		t.Errorf("DeleteSong(ctx, %q) returned an unexpected error: %s", expected.UUID, err)
	}
	if !ok {
		t.Errorf("DeleteSong(ctx, %q) returned ok = %t, expected %t", expected.UUID, ok, true)
	}

	// repeat delete to test idempotency
	ok, err = repo.DeleteSong(context.TODO(), expected.UUID)
	if err != nil {
		t.Errorf("DeleteSong(ctx, %q) [2nd time] returned an unexpected error: %s", expected.UUID, err)
	}
	if ok {
		t.Errorf("DeleteSong(ctx, %q) [2nd time] returned ok = %t, expected %t", expected.UUID, ok, false)
	}
}

func Test_fakeRepo_UpdateSong(t *testing.T) {
	t.Parallel()

	repo := NewFakeRepository()
	expected := model.Song{}
	_ = repo.CreateSong(context.TODO(), &expected)

	t.Run("found", func(t *testing.T) {
		update := model.Song{
			Model: expected.Model,
		}
		update.Title = "New"
		err := repo.UpdateSong(context.TODO(), &update)
		if err != nil {
			t.Errorf("UpdateSong(ctx, &update) returned an unexpected error: %s", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		update := model.Song{
			Model: model.Model{UUID: uuid.New()},
		}
		err := repo.UpdateSong(context.TODO(), &update)
		if !errors.Is(err, core.ErrNotFound) {
			t.Errorf("UpdateSong(ctx, &update) returned an unexpected error: %s, expected ErrNotFound", err)
		}
	})
}
