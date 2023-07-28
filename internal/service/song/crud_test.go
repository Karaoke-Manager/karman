package song

import (
	"context"
	"testing"
	"time"

	"codello.dev/ultrastar"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/internal/model"
)

func setupService(t *testing.T) (Service, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	err = db.AutoMigrate(&model.Song{}, &model.File{}, &model.Upload{})
	require.NoError(t, err)
	return NewService(db), db
}

func TestService_SaveSong(t *testing.T) {
	ctx := context.Background()
	svc, db := setupService(t)

	t.Run("uuid generation", func(t *testing.T) {
		song := model.NewSong()
		err := svc.SaveSong(ctx, &song)
		require.NoError(t, err)
		assert.NotEmpty(t, song.UUID)
	})
	t.Run("music encoding", func(t *testing.T) {
		data := ultrastar.NewSongWithBPM(120)
		data.MusicP1.AddNote(ultrastar.Note{
			Type:     ultrastar.NoteTypeRegular,
			Start:    0,
			Duration: 15,
			Pitch:    3,
			Text:     "nothing",
		})
		data.MusicP2 = nil
		song := model.NewSong()
		svc.UpdateSongFromData(&song, data)
		assert.Nil(t, song.MusicP2)
		assert.NotNil(t, song.MusicP1)
		assert.Equal(t, data.MusicP1, song.MusicP1)

		require.NoError(t, svc.SaveSong(ctx, &song))

		song, err := svc.GetSong(ctx, song.UUID)
		require.NoError(t, err)
		assert.Nil(t, song.MusicP2)
		assert.NotNil(t, song.MusicP1)
		assert.Equal(t, data.MusicP1, song.MusicP1)
	})

	t.Run("update file reference", func(t *testing.T) {
		song := model.NewSong()
		require.NoError(t, svc.SaveSong(ctx, &song))
		file := model.File{Size: 123, Type: "text/plain"}
		require.NoError(t, db.Save(&file).Error)
		song.CoverFile = &file
		require.NoError(t, svc.SaveSong(ctx, &song))
		assert.Equal(t, file.ID, *song.CoverFileID)
	})

}

func TestService_FindSongs(t *testing.T) {
	ctx := context.Background()
	svc, db := setupService(t)
	song := model.NewSong()
	song.Title = "Song 1"
	require.NoError(t, db.Save(&song).Error)
	song = model.NewSong()
	song.Title = "Song 2"
	require.NoError(t, db.Save(&song).Error)

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
	svc, db := setupService(t)

	t.Run("empty", func(t *testing.T) {
		_, err := svc.GetSong(ctx, uuid.New())
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	audio := model.File{
		Size:     1245,
		Type:     "audio/mpeg",
		Duration: 3 * time.Minute,
	}
	audio.ID = 123
	audio.UUID = uuid.New()
	expected := model.NewSong()
	expected.Title = "Hello World"
	expected.Edition = "Testing"
	expected.AudioFile = &audio
	require.NoError(t, db.Save(&expected).Error)
	t.Run("read", func(t *testing.T) {
		song, err := svc.GetSong(ctx, expected.UUID)
		assert.NoError(t, err)
		assert.Equal(t, expected.UUID, song.UUID)
		assert.Equal(t, expected.Title, song.Title)
		assert.Equal(t, expected.Edition, song.Edition)
		require.NotNil(t, song.AudioFileID)
		assert.Equal(t, uint(123), *song.AudioFileID)
		assert.Nil(t, song.AudioFile)
	})
	t.Run("include files", func(t *testing.T) {
		song, err := svc.GetSongWithFiles(ctx, expected.UUID)
		assert.NoError(t, err)
		assert.Equal(t, expected.UUID, song.UUID)
		assert.Equal(t, expected.Title, song.Title)
		assert.Equal(t, expected.Edition, song.Edition)
		require.NotNil(t, song.AudioFileID)
		assert.Equal(t, uint(123), *song.AudioFileID)
		require.NotNil(t, song.AudioFile)
		assert.Equal(t, audio.UUID, song.AudioFile.UUID)
	})
}

func TestService_DeleteSongByUUID(t *testing.T) {
	ctx := context.Background()
	svc, db := setupService(t)

	t.Run("success", func(t *testing.T) {
		song := model.NewSong()
		song.Title = "Song 1"
		require.NoError(t, db.Save(&song).Error)

		err := svc.DeleteSongByUUID(ctx, song.UUID)
		assert.NoError(t, err)
		var total int64
		require.NoError(t, db.Model(&song).Count(&total).Error)
		assert.Equal(t, int64(0), total)
	})

	t.Run("already absent", func(t *testing.T) {
		err := svc.DeleteSongByUUID(ctx, uuid.New())
		assert.NoError(t, err)
	})
}
