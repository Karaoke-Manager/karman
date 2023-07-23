package song

import (
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestService_SongData(t *testing.T) {
	svc, _ := setupService(t)

	audio := model.File{
		Size:     1234,
		Type:     "audio/mpeg",
		Bitrate:  592232,
		Duration: 3 * time.Minute,
	}
	audio.ID = 123
	audio.UUID = uuid.New()
	video := model.File{
		Size:     5823,
		Type:     "video/mp4",
		Bitrate:  1231241,
		Duration: 3 * time.Minute,
	}
	video.ID = 456
	audio.UUID = uuid.New()
	song := model.NewSong()
	song.Artist = "Foobar"
	song.Title = "Hello World"
	song.AudioFileID = &audio.ID
	song.AudioFile = &audio
	song.VideoFileID = &video.ID
	song.VideoFile = &video
	usSong := svc.SongData(song)

	assert.Equal(t, song.Artist, usSong.Artist)
	assert.Equal(t, song.Title, usSong.Title)
	assert.Equal(t, "Foobar - Hello World [AUDIO].mp3", usSong.AudioFile)
	assert.Equal(t, "Foobar - Hello World [VIDEO].m4v", usSong.VideoFile)
	assert.Empty(t, usSong.CoverFile)
	assert.Empty(t, usSong.BackgroundFile)
}
