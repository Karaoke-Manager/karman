//go:build database

package song

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	svc "github.com/Karaoke-Manager/karman/service"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func Test_dbRepo_CreateSong(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(db)
	existing := testdata.SimpleSong(t, db)

	t.Run("success", func(t *testing.T) {
		song := model.Song{}
		song.Artists = []string{"Foo", "Bar"}
		song.Title = "Hello"
		song.Genre = "World"
		err := repo.CreateSong(context.TODO(), &song)
		if err != nil {
			t.Errorf("CreateSong(ctx, &song) returned an unexpected error: %s", err)
			return
		}
		if song.UUID == uuid.Nil {
			t.Errorf("CreateSong(ctx, &song) produced song.UUID = <uuid.Nil>, expected a valid UUID")
		}
		if song.CreatedAt.IsZero() {
			t.Errorf("CreateSong(ctx, &song) produced song.CreatedAt = 0, expected a valid date")
		}
		if song.UpdatedAt.IsZero() {
			t.Errorf("CreateSong(ctx, &song) produced song.UpdatedAt = 0, expected a valid date")
		}
		if !song.DeletedAt.IsZero() {
			t.Errorf("CreateSong(ctx, &song) produced song.DeletedAt = %q, expected zero value", song.DeletedAt)
		}
	})

	t.Run("existing UUID", func(t *testing.T) {
		song := model.Song{}
		song.UUID = existing.UUID
		err := repo.CreateSong(context.TODO(), &song)
		if err != nil {
			t.Errorf("CreateSong(ctx, &song) returned an unexpected error: %s", err)
			return
		}
		if song.UUID == existing.UUID {
			t.Errorf("CreateSong(ctx, &song) did not change song.UUID, expected change")
		}
		if song.CreatedAt == existing.CreatedAt {
			t.Errorf("CreateSong(ctx, &song) did not change song.CreatedAt, expected change")
		}
	})
}

func Test_dbRepo_GetSong(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(db)
	existing := testdata.SimpleSong(t, db)

	t.Run("existing", func(t *testing.T) {
		song, err := repo.GetSong(context.TODO(), existing.UUID)
		if err != nil {
			t.Errorf("GetSong(ctx, %q) returned an unexpected error: %s", existing.UUID, err)
		}
		if song.Title != existing.Title {
			t.Errorf("song.Title = %q, expected %q", song.Title, existing.Title)
		}
	})

	t.Run("missing", func(t *testing.T) {
		id := uuid.New()
		_, err := repo.GetSong(context.TODO(), id)
		if err == nil {
			t.Errorf("GetSong(ctx, %q) did not return an error, expected ErrNotFound", id)
		} else if !errors.Is(err, svc.ErrNotFound) {
			t.Errorf("GetSong(ctx, %q) returned an unexpected error: %s", id, err)
		}
	})
}

func Test_dbRepo_FindSongs(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(db)
	testdata.NSongs(t, db, 100)

	cases := map[string]struct {
		Limit  int
		Offset int64
		Len    int
	}{
		"all":             {-1, 0, 100},
		"offset":          {-1, 2, 98},
		"offset past end": {-1, 90, 10},
		"limit zero":      {0, 0, 0},
		"limit":           {2, 0, 2},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			songs, total, err := repo.FindSongs(context.TODO(), c.Limit, c.Offset)
			if err != nil {
				t.Errorf("FindSongs(ctx, %d, %d) returned an unexpected error: %d", c.Limit, c.Offset, err)
				return
			}
			if total != 100 {
				t.Errorf("FindSongs(ctx, %d, %d) = _, %d, _, expected %d", c.Limit, c.Offset, total, 100)
			}
			if len(songs) != c.Len {
				t.Errorf("FindSongs(ctx, %d, %d) returned %d songs, expected %d", c.Limit, c.Offset, len(songs), c.Len)
			}
		})
	}
}

func Test_dbRepo_UpdateSong(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(db)

	t.Run("metadata", func(t *testing.T) {
		song := testdata.SimpleSong(t, db)
		song.Title = "Changed"
		oldUpdatedAt := song.UpdatedAt
		err := repo.UpdateSong(context.TODO(), &song)
		if err != nil {
			t.Errorf("UpdateSong(ctx, &song) returned an unexpected error: %s", err)
			return
		}
		if song.UpdatedAt == oldUpdatedAt {
			t.Errorf("UpdateSong(ctx, &song) did not change song.UpdatedAt, expected change")
		}
		if song.Title != "Changed" {
			t.Errorf("UpdateSong(ctx, &song) produced song.Title = %q, expected %q", song.Title, "Changed")
		}
	})

	t.Run("set file", func(t *testing.T) {
		song := testdata.SimpleSong(t, db)
		file := testdata.AudioFile(t, db)
		song.AudioFile = &file
		err := repo.UpdateSong(context.TODO(), &song)
		if err != nil {
			t.Errorf("UpdateSong(ctx, &song) returned an unexpected error: %s", err)
			return
		}
		if song.AudioFile == nil {
			t.Errorf("UpdateSong(ctx, &song) produced song.AudioFile = nil, expected a value")
		}
	})

	t.Run("absent file", func(t *testing.T) {
		song := testdata.SimpleSong(t, db)
		song.AudioFile = &model.File{
			Model:    model.Model{UUID: uuid.New()},
			Type:     mediatype.AudioMPEG,
			Duration: 3 * time.Minute,
		}
		err := repo.UpdateSong(context.TODO(), &song)
		if err != nil {
			t.Errorf("UpdateSong(ctx, &song) returned an unexpected error: %s", err)
			return
		}
		if song.AudioFile != nil {
			t.Errorf("UpdateSong(ctx, &song) did not clear song.AudioFile, expected no audio file")
		}
	})
}

func Test_dbRepo_DeleteSong(t *testing.T) {
	t.Parallel()

	db := test.NewDB(t)
	repo := NewDBRepository(db)
	song := testdata.SimpleSong(t, db)

	ok, err := repo.DeleteSong(context.TODO(), song.UUID)
	if err != nil {
		t.Errorf("DeleteSong(ctx, %q) returned an unexpected error: %s", song.UUID, err)
		return
	}
	if !ok {
		t.Errorf("DeleteSong(ctx, %q) = %t, _, expected %t", song.UUID, ok, true)
	}
	// repeat delete to test idempotency
	ok, err = repo.DeleteSong(context.TODO(), song.UUID)
	if err != nil {
		t.Errorf("DeleteSong(ctx, %q) [2nd time] returned an unexpected error: %s", song.UUID, err)
		return
	}
	if ok {
		t.Errorf("DeleteSong(ctx, %q) [2nd time] = %t, _, expected %t", song.UUID, ok, false)
	}
}
