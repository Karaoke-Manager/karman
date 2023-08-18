package test

import (
	"encoding/hex"
	"time"

	"codello.dev/ultrastar"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/service/entity"
)

// A Dataset provides named values for the expected content of a testing database.
type Dataset struct {
	InvalidUUID string

	TotalSongs   int64
	TotalUploads int64

	ImageFile *model.File // may be used by multiple songs
	AudioFile *model.File // may be used by multiple songs
	VideoFile *model.File // may be used by multiple songs

	AbsentUploadUUID uuid.UUID     // may be a present UUID for other types
	OpenUpload       *model.Upload // files can be uploaded
	PendingUpload    *model.Upload // enqueued for processing
	ProcessingUpload *model.Upload // currently being processed
	UploadWithSongs  *model.Upload // done processing
	UploadWithErrors *model.Upload

	AbsentSongUUID uuid.UUID   // may be a present UUID for other types
	AbsentSong     *model.Song // UUID is AbsentSongUUID
	BasicSong      *model.Song // no media, no upload, no music

	SongWithUpload     *model.Song
	SongWithCover      *model.Song // may or may not have other media
	SongWithBackground *model.Song // may or may not have other media
	SongWithAudio      *model.Song // may or may not have other media
	SongWithVideo      *model.Song // may or may not have other media
}

// NewDataset creates a new Dataset and stores it into the db.
func NewDataset(db *gorm.DB) *Dataset {
	data := &Dataset{
		InvalidUUID:      "Hello%20World",
		AbsentSongUUID:   uuid.New(),
		AbsentUploadUUID: uuid.New(),
		TotalUploads:     5,
	}

	song := entity.Song{Entity: entity.Entity{UUID: data.AbsentSongUUID}}
	data.AbsentSong = song.ToModel()

	checksum, _ := hex.DecodeString("d2a84f4b8b650937ec8f73cd8be2c74add5a911ba64df27458ed8229da804a26")
	imageFile := entity.File{
		Type:     mediatype.ImagePNG,
		Size:     1235,
		Checksum: checksum,
		Width:    512,
		Height:   512,
	}
	db.Save(&imageFile)
	data.ImageFile = imageFile.ToModel()

	audioFile := entity.File{
		Type:     mediatype.AudioMPEG,
		Size:     62352,
		Checksum: checksum,
		Duration: 3 * time.Minute,
	}
	db.Save(&audioFile)
	data.AudioFile = audioFile.ToModel()

	videoFile := entity.File{
		Type:     mediatype.VideoMP4,
		Size:     123151,
		Checksum: checksum,
		Duration: 2 * time.Second,
	}
	db.Save(&videoFile)
	data.VideoFile = videoFile.ToModel()

	upload := entity.Upload{
		Open:           true,
		SongsTotal:     -1,
		SongsProcessed: -1,
	}
	db.Save(&upload)
	data.OpenUpload = upload.ToModel()

	upload = entity.Upload{
		Open:           false,
		SongsTotal:     -1,
		SongsProcessed: -1,
	}
	db.Save(&upload)
	data.PendingUpload = upload.ToModel()

	upload = entity.Upload{
		Open:           false,
		SongsTotal:     -1,
		SongsProcessed: 0,
	}
	db.Save(&upload)
	data.ProcessingUpload = upload.ToModel()

	upload = entity.Upload{
		Open:           false,
		SongsTotal:     0,
		SongsProcessed: 0,
		ProcessingErrors: []entity.UploadProcessingError{
			{File: "file1.txt", Message: "could not parse"},
			{File: "file2.txt", Message: "could not read"},
		},
	}
	db.Save(&upload)
	data.UploadWithErrors = upload.ToModel()

	uploadWithSongs := entity.Upload{
		Open:             false,
		SongsTotal:       4,
		SongsProcessed:   4,
		ProcessingErrors: nil,
	}
	db.Save(&uploadWithSongs)
	data.UploadWithSongs = uploadWithSongs.ToModel()

	song = entity.Song{
		Title:    "Cold",
		Artist:   "Darrin DuBuque",
		Genre:    "Latin",
		Language: "English",
		Year:     2003,
	}
	db.Save(&song)
	data.BasicSong = song.ToModel()

	song = entity.Song{
		Upload:   &uploadWithSongs,
		Title:    "More",
		Artist:   "Nobory",
		Genre:    "Rock",
		Language: "English",
	}
	db.Save(&song)
	data.SongWithUpload = song.ToModel()

	song = entity.Song{
		Title:     "Some",
		Artist:    "Unimportant",
		CoverFile: &imageFile,
		MusicP1:   ultrastar.NewMusic(),
	}
	db.Save(&song)
	data.SongWithCover = song.ToModel()

	song = entity.Song{
		Title:          "Whatever",
		Edition:        "SingStar",
		BackgroundFile: &imageFile,
	}
	db.Save(&song)
	data.SongWithBackground = song.ToModel()

	song = entity.Song{
		Title:     "Whatever",
		Gap:       1252,
		AudioFile: &audioFile,
	}
	db.Save(&song)
	data.SongWithAudio = song.ToModel()

	song = entity.Song{
		Title:     "Whatever",
		Comment:   "useless",
		VideoFile: &videoFile,
	}
	db.Save(&song)
	data.SongWithVideo = song.ToModel()

	data.TotalSongs = 150
	for i := int64(0); i < data.TotalSongs-5; i++ {
		// Some dummy data
		song := entity.Song{}
		db.Save(&song)
	}
	return data
}
