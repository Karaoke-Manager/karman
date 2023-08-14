package song

import (
	"context"
	"testing"

	"codello.dev/ultrastar"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/internal/test"
)

func setupService(t *testing.T, withData bool) (s Service, data *test.Dataset) {
	db := test.NewDB(t)
	if withData {
		data = test.NewDataset(db)
	}
	s = NewService(db)
	return
}

func TestService_CreateSong(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t, false)

	t.Run("uuid generation", func(t *testing.T) {
		song := &model.Song{}
		err := svc.CreateSong(ctx, song)
		require.NoError(t, err)
		assert.NotEmpty(t, song.UUID)
	})
	t.Run("music encoding", func(t *testing.T) {
		song := &model.Song{
			Song: *ultrastar.NewSongWithBPM(120),
		}
		song.MusicP1.AddNote(ultrastar.Note{
			Type:     ultrastar.NoteTypeRegular,
			Start:    0,
			Duration: 15,
			Pitch:    3,
			Text:     "nothing",
		})
		if assert.NoError(t, svc.CreateSong(ctx, song)) {
			song2, err := svc.GetSong(ctx, song.UUID)
			require.NoError(t, err)
			assert.Nil(t, song2.MusicP2)
			assert.NotNil(t, song2.MusicP1)
			assert.Equal(t, song.MusicP1, song2.MusicP1)
		}
	})
}

func TestService_UpdateSongData(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)
	song := &model.Song{}
	*song = *data.BasicSong
	song.Title = "Foobar"
	song.Artist = "Hey"
	err := svc.UpdateSongData(ctx, song)
	assert.NoError(t, err)

	song, err = svc.GetSong(ctx, data.BasicSong.UUID)
	if assert.NoError(t, err) {
		assert.Equal(t, "Foobar", song.Title)
		assert.Equal(t, "Hey", song.Artist)
	}
}

func TestService_FindSongs(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	cases := map[string]struct {
		limit          int
		offset         int
		expectedLength int
	}{
		"no offset":     {25, 0, 25},
		"with offset":   {10, 17, 10},
		"cutoff offset": {40, int(data.TotalSongs) - 20, 20},
		"past end":      {8, int(data.TotalSongs) + 10, 0},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			songs, total, err := svc.FindSongs(ctx, c.limit, c.offset)
			assert.NoError(t, err)
			assert.Equal(t, data.TotalSongs, total)
			assert.Len(t, songs, c.expectedLength)
		})
	}
}

func TestService_GetSong(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	t.Run("empty", func(t *testing.T) {
		_, err := svc.GetSong(ctx, data.AbsentSongUUID)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("read", func(t *testing.T) {
		song, err := svc.GetSong(ctx, data.SongWithAudio.UUID)
		if assert.NoError(t, err) {
			assert.Equal(t, data.SongWithAudio.Title, song.Title)
			assert.Equal(t, data.SongWithAudio.Artist, song.Artist)
			assert.Equal(t, data.SongWithAudio.MusicP1, song.MusicP1)
			assert.Equal(t, data.SongWithAudio.AudioFile.Duration, song.AudioFile.Duration)
		}
	})

	t.Run("file names", func(t *testing.T) {
		song, err := svc.GetSong(ctx, data.SongWithVideo.UUID)
		if assert.NoError(t, err) {
			assert.Empty(t, song.CoverFileName)
			assert.NotEmpty(t, song.VideoFileName)
		}
	})
}

func TestService_DeleteSong(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	t.Run("success", func(t *testing.T) {
		err := svc.DeleteSong(ctx, data.BasicSong.UUID)
		assert.NoError(t, err)
		_, err = svc.GetSong(ctx, data.BasicSong.UUID)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("already absent", func(t *testing.T) {
		err := svc.DeleteSong(ctx, uuid.New())
		assert.NoError(t, err)
	})
}
