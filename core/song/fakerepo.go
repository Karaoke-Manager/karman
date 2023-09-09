package song

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/model"
)

// fakeRepo is a simple implementation of Repository that can be used for testing.
type fakeRepo struct {
	// songs is the "database" of a fakeRepo.
	songs map[uuid.UUID]model.Song
}

// NewFakeRepository returns a new Repository implementation backed by an in-memory map.
// Not all features (especially complex queries) are supported on fake repositories.
func NewFakeRepository() Repository {
	return &fakeRepo{make(map[uuid.UUID]model.Song)}
}

// CreateSong stores the song and sets its UUID, CreatedAt, and UpdatedAt fields.
func (r *fakeRepo) CreateSong(_ context.Context, song *model.Song) error {
	song.UUID = uuid.New()
	song.CreatedAt = time.Now()
	song.UpdatedAt = song.CreatedAt
	r.songs[song.UUID] = *song
	return nil
}

// GetSong fetches looks up the song with the specified UUID.
func (r *fakeRepo) GetSong(_ context.Context, id uuid.UUID) (model.Song, error) {
	song, ok := r.songs[id]
	if !ok {
		return model.Song{}, core.ErrNotFound
	}
	return song, nil
}

// FindSongs returns a list of songs limited by the specified pagination parameters.
// This implementation does not support complex filter queries.
func (r *fakeRepo) FindSongs(_ context.Context, limit int, offset int64) ([]model.Song, int64, error) {
	if limit < 0 {
		limit = math.MaxInt
	}
	songs := make([]model.Song, 0)
	idx := int64(0)
	for _, song := range r.songs {
		if idx < offset {
			idx++
			continue
		}
		if len(songs) >= limit {
			break
		}
		songs = append(songs, song)
	}
	return songs, int64(len(r.songs)), nil
}

// DeleteSong deletes the song with the specified UUID (if it exists).
func (r *fakeRepo) DeleteSong(_ context.Context, id uuid.UUID) (bool, error) {
	_, ok := r.songs[id]
	delete(r.songs, id)
	return ok, nil
}

// UpdateSong updates the data of song.
func (r *fakeRepo) UpdateSong(_ context.Context, song *model.Song) error {
	_, ok := r.songs[song.UUID]
	if !ok {
		return core.ErrNotFound
	}
	song.UpdatedAt = time.Now()
	r.songs[song.UUID] = *song
	return nil
}
