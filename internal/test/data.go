package test

import (
	"codello.dev/ultrastar"
	"encoding/hex"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"

	"github.com/Karaoke-Manager/karman/internal/model"
)

// A Dataset provides named values for the expected content of a testing database.
type Dataset struct {
	InvalidUUID string

	ImageFile model.File // may be used by multiple songs
	AudioFile model.File // may be used by multiple songs
	VideoFile model.File // may be used by multiple songs

	UploadWithSongs model.Upload

	AbsentSongUUID uuid.UUID  // may be a present UUID for other types
	AbsentSong     model.Song // UUID is AbsentSongUUID
	BasicSong      model.Song // no media, no upload, no music

	SongWithUpload     model.Song
	SongWithCover      model.Song // may or may not have other media
	SongWithBackground model.Song // may or may not have other media
	SongWithAudio      model.Song // may or may not have other media
	SongWithVideo      model.Song // may or may not have other media
}

// NewDataset creates a new Dataset and stores it into the db.
func NewDataset(db *gorm.DB) *Dataset {
	data := &Dataset{
		InvalidUUID:    "Hello%20World",
		AbsentSongUUID: uuid.New(),
	}

	data.AbsentSong = model.Song{Model: model.Model{UUID: data.AbsentSongUUID}}

	checksum, _ := hex.DecodeString("d2a84f4b8b650937ec8f73cd8be2c74add5a911ba64df27458ed8229da804a26")
	data.ImageFile = model.File{
		Type:     mediatype.ImagePNG,
		Size:     1235,
		Checksum: checksum,
		Width:    512,
		Height:   512,
	}
	db.Save(&data.ImageFile)

	data.AudioFile = model.File{
		Type:     mediatype.AudioMPEG,
		Size:     62352,
		Checksum: checksum,
		Duration: 3 * time.Minute,
	}
	db.Save(&data.AudioFile)

	data.VideoFile = model.File{
		Type:     mediatype.VideoMP4,
		Size:     123151,
		Checksum: checksum,
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

	data.BasicSong = model.Song{
		Title:      "Cold",
		Artist:     "Darrin DuBuque",
		Genre:      "Latin",
		Language:   "English",
		Year:       2003,
		CalcMedley: true,
	}
	db.Save(&data.BasicSong)

	data.SongWithUpload = model.Song{
		Upload:     &data.UploadWithSongs,
		Title:      "More",
		Artist:     "Nobory",
		Genre:      "Rock",
		Language:   "English",
		CalcMedley: true,
	}
	db.Save(&data.SongWithUpload)

	data.SongWithCover = model.Song{
		Title:      "Some",
		Artist:     "Unimportant",
		CoverFile:  &data.ImageFile,
		CalcMedley: true,
		MusicP1:    ultrastar.NewMusic(),
	}
	db.Save(&data.SongWithCover)

	data.SongWithBackground = model.Song{
		Title:          "Whatever",
		Edition:        "SingStar",
		BackgroundFile: &data.ImageFile,
		CalcMedley:     true,
	}
	db.Save(&data.SongWithBackground)

	data.SongWithAudio = model.Song{
		Title:      "Whatever",
		Gap:        1252,
		AudioFile:  &data.AudioFile,
		CalcMedley: true,
	}
	db.Save(&data.SongWithAudio)

	data.SongWithVideo = model.Song{
		Title:      "Whatever",
		Comment:    "useless",
		VideoFile:  &data.VideoFile,
		CalcMedley: true,
	}
	db.Save(&data.SongWithVideo)

	for i := 0; i < 145; i++ {
		// Some dummy data
		song := model.NewSong()
		db.Save(&song)
	}
	return data
}
