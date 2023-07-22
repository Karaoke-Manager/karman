package song

import (
	"context"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
)

func setupService(t *testing.T) Service {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	err = db.AutoMigrate(&model.Song{}, &model.File{}, &model.Upload{})
	require.NoError(t, err)
	return NewService(db)
}

func TestService_CreateSong(t *testing.T) {
	ctx := context.Background()
	svc := setupService(t)

	t.Run("uuid generation", func(t *testing.T) {
		song := model.NewSong()
		song.UUID = uuid.Nil
		err := svc.SaveSong(ctx, &song)
		require.NoError(t, err)
		assert.NotEmpty(t, song.UUID)
	})
	t.Run("music encoding", func(t *testing.T) {
		song := ultrastar.NewSongWithBPM(120)
		song.MusicP1.AddNote(ultrastar.Note{
			Type:     ultrastar.NoteTypeRegular,
			Start:    0,
			Duration: 15,
			Pitch:    3,
			Text:     "nothing",
		})
		song.MusicP2 = nil
		s, err := svc.CreateSong(ctx, song)
		require.NoError(t, err)
		assert.Nil(t, s.MusicP2)
		assert.NotNil(t, s.MusicP1)
		assert.Equal(t, song.MusicP1, s.MusicP1)

		s, err = svc.GetSong(ctx, s.UUID.String())
		require.NoError(t, err)
		assert.Nil(t, s.MusicP2)
		assert.NotNil(t, s.MusicP1)
		assert.Equal(t, song.MusicP1, s.MusicP1)
	})
}

func TestService_FindSongs(t *testing.T) {
	ctx := context.Background()
	svc := setupService(t)
	song := model.NewSong()
	song.Title = "Song 1"
	require.NoError(t, svc.SaveSong(ctx, &song))
	song = model.NewSong()
	song.Title = "Song 2"
	require.NoError(t, svc.SaveSong(ctx, &song))

	t.Run("all songs", func(t *testing.T) {
		songs, total, err := svc.FindSongs(ctx, 25, 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, songs, 2)
	})
	t.Run("with offset", func(t *testing.T) {
		songs, total, err := svc.FindSongs(ctx, 25, 1)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, songs, 1)
	})
	t.Run("large offset", func(t *testing.T) {
		songs, total, err := svc.FindSongs(ctx, 25, 3)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, songs, 0)
	})
}

func TestService_GetSong(t *testing.T) {
	ctx := context.Background()
	svc := setupService(t)

	t.Run("empty", func(t *testing.T) {
		_, err := svc.GetSong(ctx, "non existing")
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	audio := model.File{
		Size:     1245,
		Type:     "audio/mpeg",
		Bitrate:  62372,
		Duration: 3 * time.Minute,
	}
	audio.ID = 123
	audio.UUID = uuid.New()
	expected := model.NewSong()
	expected.Title = "Hello World"
	expected.Edition = "Testing"
	expected.AudioFile = &audio
	err := svc.SaveSong(ctx, &expected)
	require.NoError(t, err)
	t.Run("read", func(t *testing.T) {
		song, err := svc.GetSong(ctx, expected.UUID.String())
		assert.NoError(t, err)
		assert.Equal(t, expected.UUID, song.UUID)
		assert.Equal(t, expected.Title, song.Title)
		assert.Equal(t, expected.Edition, song.Edition)
		require.NotNil(t, song.AudioFileID)
		assert.Equal(t, uint(123), *song.AudioFileID)
		assert.Nil(t, song.AudioFile)
	})
	t.Run("include files", func(t *testing.T) {
		song, err := svc.GetSongWithFiles(ctx, expected.UUID.String())
		assert.NoError(t, err)
		assert.Equal(t, expected.UUID, song.UUID)
		assert.Equal(t, expected.Title, song.Title)
		assert.Equal(t, expected.Edition, song.Edition)
		require.NotNil(t, song.AudioFileID)
		assert.Equal(t, uint(123), *song.AudioFileID)
		require.NotNil(t, song.AudioFile)
		assert.Equal(t, audio.UUID, song.AudioFile.UUID)
		assert.Equal(t, audio.Bitrate, song.AudioFile.Bitrate)
	})
}

func TestService_DeleteSongByUUID(t *testing.T) {
	ctx := context.Background()
	svc := setupService(t)
	song := model.NewSong()
	song.UUID = uuid.New()
	song.Title = "Song 1"
	require.NoError(t, svc.SaveSong(ctx, &song))

	err := svc.DeleteSongByUUID(ctx, song.UUID.String())
	assert.NoError(t, err)
	_, total, err := svc.FindSongs(ctx, 25, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
}
