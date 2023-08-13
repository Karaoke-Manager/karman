package song

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_ReplaceCover(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	t.Run("set cover", func(t *testing.T) {
		err := svc.ReplaceCover(ctx, data.BasicSong, data.ImageFile)
		assert.NoError(t, err)
		assert.Equal(t, data.ImageFile, data.BasicSong.CoverFile)
		song, err := svc.GetSong(ctx, data.BasicSong.UUID)
		if assert.NoError(t, err) {
			assert.Equal(t, data.ImageFile, song.CoverFile)
		}
	})
	t.Run("delete cover", func(t *testing.T) {
		err := svc.ReplaceCover(ctx, data.SongWithCover, nil)
		assert.NoError(t, err)
		assert.Nil(t, data.SongWithCover.CoverFile)
		song, err := svc.GetSong(ctx, data.SongWithCover.UUID)
		if assert.NoError(t, err) {
			assert.Nil(t, song.CoverFile)
		}
	})
}

func TestService_ReplaceAudio(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	t.Run("set audio", func(t *testing.T) {
		err := svc.ReplaceAudio(ctx, data.BasicSong, data.AudioFile)
		assert.NoError(t, err)
		assert.Equal(t, data.AudioFile, data.BasicSong.AudioFile)
		song, err := svc.GetSong(ctx, data.BasicSong.UUID)
		if assert.NoError(t, err) {
			assert.Equal(t, data.AudioFile, song.AudioFile)
		}
	})
	t.Run("delete audio", func(t *testing.T) {
		err := svc.ReplaceAudio(ctx, data.SongWithAudio, nil)
		assert.NoError(t, err)
		assert.Nil(t, data.SongWithAudio.AudioFile)
		song, err := svc.GetSong(ctx, data.SongWithAudio.UUID)
		if assert.NoError(t, err) {
			assert.Nil(t, song.AudioFile)
		}
	})
}

func TestService_ReplaceBackground(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	t.Run("set background", func(t *testing.T) {
		err := svc.ReplaceBackground(ctx, data.BasicSong, data.ImageFile)
		assert.NoError(t, err)
		assert.Equal(t, data.ImageFile, data.BasicSong.BackgroundFile)
		song, err := svc.GetSong(ctx, data.BasicSong.UUID)
		if assert.NoError(t, err) {
			assert.Equal(t, data.ImageFile, song.BackgroundFile)
		}
	})
	t.Run("delete background", func(t *testing.T) {
		err := svc.ReplaceBackground(ctx, data.SongWithBackground, nil)
		assert.NoError(t, err)
		assert.Nil(t, data.SongWithBackground.BackgroundFile)
		song, err := svc.GetSong(ctx, data.SongWithBackground.UUID)
		if assert.NoError(t, err) {
			assert.Nil(t, song.BackgroundFile)
		}
	})
}

func TestService_ReplaceVideo(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	t.Run("set video", func(t *testing.T) {
		err := svc.ReplaceVideo(ctx, data.BasicSong, data.VideoFile)
		assert.NoError(t, err)
		assert.Equal(t, data.VideoFile, data.BasicSong.VideoFile)
		song, err := svc.GetSong(ctx, data.BasicSong.UUID)
		if assert.NoError(t, err) {
			assert.Equal(t, data.VideoFile, song.VideoFile)
		}
	})
	t.Run("delete video", func(t *testing.T) {
		err := svc.ReplaceVideo(ctx, data.SongWithVideo, nil)
		assert.NoError(t, err)
		assert.Nil(t, data.SongWithVideo.VideoFile)
		song, err := svc.GetSong(ctx, data.SongWithVideo.UUID)
		if assert.NoError(t, err) {
			assert.Nil(t, song.VideoFile)
		}
	})
}
