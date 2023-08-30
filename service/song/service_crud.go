package song

import (
	"context"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
)

// FindSongs fetches a page of songs from the database.
func (s *service) FindSongs(ctx context.Context, limit int, offset int64) ([]*model.Song, int64, error) {
	songs, total, err := s.db.FindSongs(ctx, limit, offset)
	for _, song := range songs {
		s.ensureFilenames(song)
		s.ensureArtists(song)
	}
	return songs, total, err
}

// GetSong fetches a single song from the database.
func (s *service) GetSong(ctx context.Context, id uuid.UUID) (*model.Song, error) {
	song, err := s.db.GetSong(ctx, id)
	if err != nil {
		return song, err
	}
	s.ensureFilenames(song)
	s.ensureArtists(song)
	return song, nil
}

// CreateSong persists a new song into the database.
// If song already has a UUID it is ignored and a new one will be assigned.
func (s *service) CreateSong(ctx context.Context, song *model.Song) error {
	return s.db.CreateSong(ctx, song)
}

// SaveSong updates song in the database.
// song must already have been persisted before or an error will be returned.
func (s *service) SaveSong(ctx context.Context, song *model.Song) error {
	return s.db.UpdateSong(ctx, song)
}

// DeleteSong deletes the song with the specified UUID from the database.
func (s *service) DeleteSong(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.DeleteSongByUUID(ctx, id)
	return err
}
