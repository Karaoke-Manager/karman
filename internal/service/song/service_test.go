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
	t.Run("create and read", func(t *testing.T) {
		song := model.NewSong()
		song.Title = "Hello World"
		song.Edition = "Testing"
		err := svc.SaveSong(ctx, &song)
		require.NoError(t, err)

		song2, err := svc.GetSong(ctx, song.UUID.String())
		assert.NoError(t, err)
		assert.Equal(t, song.UUID, song2.UUID)
		assert.Equal(t, song.Title, song2.Title)
		assert.Equal(t, song.Edition, song2.Edition)
	})
}

func TestRepository_DeleteSongByUUID(t *testing.T) {
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
