package test

import (
	"codello.dev/ultrastar"
	"encoding/hex"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"

	"github.com/Karaoke-Manager/karman/internal/model"
)

type Dataset struct {
	InvalidUUID string

	ImageFile model.File
	AudioFile model.File
	VideoFile model.File

	UploadWithSongs model.Upload

	AbsentSongUUID           uuid.UUID
	AbsentSong               model.Song
	SongWithoutMediaAndMusic model.Song

	SongWithUpload     model.Song
	SongWithCover      model.Song
	SongWithBackground model.Song
	SongWithAudio      model.Song
	SongWithVideo      model.Song
}

func NewDataset(db *gorm.DB) *Dataset {
	data := &Dataset{
		InvalidUUID:    "Hello%20World",
		AbsentSongUUID: uuid.New(),
	}

	data.AbsentSong = model.Song{Model: model.Model{UUID: data.AbsentSongUUID}}

	checksum, _ := hex.DecodeString("d2a84f4b8b650937ec8f73cd8be2c74add5a911ba64df27458ed8229da804a26")
	data.ImageFile = model.File{
		Type:     "image/png",
		Size:     1235,
		Checksum: checksum,
		Width:    512,
		Height:   512,
	}
	db.Save(&data.ImageFile)

	data.AudioFile = model.File{
		Type:     "audio/mpeg",
		Size:     62352,
		Checksum: checksum,
		Bitrate:  128000,
		Duration: 3 * time.Minute,
	}
	db.Save(&data.AudioFile)

	data.VideoFile = model.File{
		Type:     "video/mp4",
		Size:     123151,
		Checksum: checksum,
		Bitrate:  5212327,
		Duration: 2 * time.Second,
	}
	db.Save(&data.VideoFile)

	data.UploadWithSongs = model.Upload{
		Open:             false,
		SongsTotal:       4,
		SongsProcessed:   3,
		ProcessingErrors: nil,
	}
	db.Save(&data.UploadWithSongs)

	data.SongWithoutMediaAndMusic = model.Song{
		Title:    "Cold",
		Artist:   "Darrin DuBuque",
		Genre:    "Latin",
		Language: "English",
		Year:     2003,
	}
	db.Save(&data.SongWithoutMediaAndMusic)

	data.SongWithUpload = model.Song{
		Upload:   &data.UploadWithSongs,
		Title:    "More",
		Artist:   "Nobory",
		Genre:    "Rock",
		Language: "English",
	}
	db.Save(&data.SongWithUpload)

	data.SongWithCover = model.Song{
		Title:     "Some",
		Artist:    "Unimportant",
		CoverFile: &data.ImageFile,
		MusicP1:   ultrastar.NewMusic(),
	}
	db.Save(&data.SongWithCover)

	data.SongWithBackground = model.Song{
		Title:          "Whatever",
		Edition:        "SingStar",
		BackgroundFile: &data.ImageFile,
	}
	db.Save(&data.SongWithBackground)

	data.SongWithAudio = model.Song{
		Title:     "Whatever",
		Gap:       1252,
		AudioFile: &data.AudioFile,
	}
	db.Save(&data.SongWithAudio)

	data.SongWithVideo = model.Song{
		Title:     "Whatever",
		Comment:   "useless",
		VideoFile: &data.VideoFile,
	}
	db.Save(&data.SongWithVideo)

	for i := 0; i < 145; i++ {
		// Some dummy data
		song := model.NewSong()
		db.Save(&song)
	}
	return data
}
