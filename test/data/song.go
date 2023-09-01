//go:build database

package testdata

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"codello.dev/ultrastar/txt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/test"
)

// readSong is a helper function that reads the named txt file from the testdata folder as a model.Song.
func readSong(t *testing.T, name string) model.Song {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "testdata", name)
	f := test.MustOpen(t, path)
	s, err := txt.NewReader(f).ReadSong()
	if err != nil {
		t.Fatalf("could not read test file at testdata/%s: %s", name, err)
	}
	return model.Song{Song: s, Artists: []string{s.Artist}}
}

// insertSong inserts the song into the database.
// You can specify additional column values via the extra map.
func insertSong(db pgxutil.DB, song *model.Song, extra map[string]any) error {
	values := map[string]any{
		"title":    song.Title,
		"artists":  song.Artists,
		"language": song.Language,
		"genre":    song.Genre,
		"year":     song.Year,
		"creator":  song.Creator,
		"bpm":      song.BPM,
		"gap":      song.Gap,
	}
	for key, value := range extra {
		values[key] = value
	}
	row, err := pgxutil.InsertRowReturning(context.TODO(), db, "songs",
		values, "id, uuid, created_at, updated_at", pgx.RowToStructByName[creationResult])
	if err != nil {
		return err
	}
	song.UUID = row.UUID
	song.CreatedAt = row.CreatedAt
	song.UpdatedAt = row.UpdatedAt
	return nil
}

// SimpleSong inserts a single song into the database and returns the inserted data.
//
//go:generate go run ../../tools/gensong -output testdata/simple-song.txt
func SimpleSong(t *testing.T, db pgxutil.DB) model.Song {
	song := readSong(t, "simple-song.txt")
	if err := insertSong(db, &song, nil); err != nil {
		t.Fatalf("test.SimpleSong() could not insert into the database: %s", err)
	}
	return song
}

// NSongs inserts n empty songs into the database.
// These should not be tested for their contents but only their existence.
func NSongs(t *testing.T, db pgxutil.DB, n int) {
	_, err := db.CopyFrom(context.TODO(), pgx.Identifier{"songs"}, []string{"title"}, pgx.CopyFromSlice(n, func(i int) ([]any, error) {
		return []any{fmt.Sprintf("Song %d", i)}, nil
	}))
	if err != nil {
		t.Fatalf("test.NSongs() could not insert all songs: %s", err)
	}
}

// SongWithUpload inserts an upload into the database that contains a single song.
// That song is returned.
//
//go:generate go run ../../tools/gensong -output testdata/song-with-upload.txt
func SongWithUpload(t *testing.T, db pgxutil.DB) model.Song {
	id, err := insertUpload(db, &model.Upload{State: model.UploadStateOpen}, nil)
	if err != nil {
		t.Fatalf("testdata.SongWithUpload() could not insert upload into the database: %s", err)
	}
	song := readSong(t, "song-with-upload.txt")
	if err := insertSong(db, &song, map[string]any{
		"upload_id": id,
	}); err != nil {
		t.Fatalf("testdata.SongWithUpload() could not insert song into the database: %s", err)
	}
	song.InUpload = true
	return song
}

//go:generate go run ../../tools/gensong -output testdata/song-with-media.txt

// SongWithAudio inserts an audio file as well as a song with that file as audio into the database.
// The song is returned.
func SongWithAudio(t *testing.T, db pgxutil.DB) model.Song {
	song := readSong(t, "song-with-media.txt")
	song.AudioFile = &model.File{
		Type:     mediatype.AudioMPEG,
		Size:     627,
		Duration: 2 * time.Minute,
	}
	audioID, err := insertFile(db, song.AudioFile, nil)
	if err != nil {
		t.Fatalf("testdata.SongWithAudio() could not insert file into the database: %s", err)
	}
	if err = insertSong(db, &song, map[string]any{
		"audio_file_id": audioID,
	}); err != nil {
		t.Fatalf("testdata.SongWithAudio() could not insert song into the database: %s", err)
	}
	return song
}

// SongWithCover inserts an image file as well as a song with that file as cover into the database.
// The song is returned.
func SongWithCover(t *testing.T, db pgxutil.DB) model.Song {
	song := readSong(t, "song-with-media.txt")
	song.CoverFile = &model.File{
		Type:   mediatype.ImageGIF,
		Size:   9962,
		Width:  512,
		Height: 512,
	}
	coverID, err := insertFile(db, song.CoverFile, nil)
	if err != nil {
		t.Fatalf("testdata.SongWithCover() could not insert file into the database: %s", err)
	}
	if err = insertSong(db, &song, map[string]any{
		"cover_file_id": coverID,
	}); err != nil {
		t.Fatalf("testdata.SongWithCover() could not insert song into the database: %s", err)
	}
	return song
}

// SongWithVideo inserts a video file as well as a song with that file as video into the database.
// The song is returned.
func SongWithVideo(t *testing.T, db pgxutil.DB) model.Song {
	song := readSong(t, "song-with-media.txt")
	song.VideoFile = &model.File{
		Type:     mediatype.VideoMP4,
		Size:     9962,
		Duration: 5 * time.Minute,
		Width:    1498,
		Height:   720,
	}
	videoID, err := insertFile(db, song.VideoFile, nil)
	if err != nil {
		t.Fatalf("testdata.SongWithVideo() could not insert file into the database: %s", err)
	}
	if err = insertSong(db, &song, map[string]any{
		"video_file_id": videoID,
	}); err != nil {
		t.Fatalf("testdata.SongWithVideo() could not insert song into the database: %s", err)
	}
	return song
}

// SongWithBackground inserts an image file as well as a song with that file as background into the database.
// The song is returned.
func SongWithBackground(t *testing.T, db pgxutil.DB) model.Song {
	song := readSong(t, "song-with-media.txt")
	song.BackgroundFile = &model.File{
		Type:   mediatype.ImageJPEG,
		Size:   99622,
		Width:  1920,
		Height: 1080,
	}
	bgID, err := insertFile(db, song.BackgroundFile, nil)
	if err != nil {
		t.Fatalf("testdata.SongWithBackground() could not insert file into the database: %s", err)
	}
	if err = insertSong(db, &song, map[string]any{
		"background_file_id": bgID,
	}); err != nil {
		t.Fatalf("testdata.SongWithBackground() could not insert song into the database: %s", err)
	}
	return song
}
