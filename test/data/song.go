//go:build database

package testdata

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"codello.dev/ultrastar/txt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/test"
)

//go:generate go run ../../tools/gensong -output testdata/simple-song.txt

// SimpleSong inserts a single song into the database and returns the inserted data.
func SimpleSong(t *testing.T, db pgxutil.DB) model.Song {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "testdata", "simple-song.txt")
	f := test.MustOpen(t, path)
	s, err := txt.NewReader(f).ReadSong()
	song := model.Song{Song: s, Artists: []string{s.Artist}}
	if err != nil {
		t.Fatalf("SimpleSong() could not read song: %s", err)
	}
	row, err := pgxutil.InsertRowReturning(context.TODO(), db, "songs", map[string]any{
		"title":    song.Title,
		"genre":    song.Genre,
		"language": song.Language,
		"year":     song.Year,
	}, "id, uuid, created_at, updated_at", pgx.RowToStructByName[creationResult])
	if err != nil {
		t.Fatalf("test.SimpleSong() could not insert into the database: %s", err)
	}
	song.UUID = row.UUID
	song.CreatedAt = row.CreatedAt
	song.UpdatedAt = row.UpdatedAt
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
