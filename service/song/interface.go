package song

import (
	"context"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
)

// A Repository is an interface for storing songs.
type Repository interface {
	// CreateSong creates a new song with the specified data.
	// An existing song.UUID must be ignored.
	// This method must set song.UUID, song.CreatedAt, and song.UpdatedAt appropriately.
	CreateSong(ctx context.Context, song *model.Song) error

	// GetSong fetches the song with the specified UUID.
	// If no such song exists, service.ErrNotFound will be returned.
	GetSong(ctx context.Context, id uuid.UUID) (model.Song, error)

	// FindSongs returns all songs matching the specified query.
	// Results are paginated with limit and offset.
	// The second return value contains the total (unpaginated) number of songs.
	//
	// If no songs match, no error will be returned.
	FindSongs(ctx context.Context, limit int, offset int64) ([]model.Song, int64, error)

	// UpdateSong saves Updates for the specified song.
	// The song's UUID must already exist in the database, otherwise e service.ErrNotFound
	// will be returned.
	UpdateSong(ctx context.Context, song *model.Song) error

	// DeleteSong deletes the song with the specified UUID.
	// If no such song exists, the first return value will be false.
	DeleteSong(ctx context.Context, id uuid.UUID) (bool, error)
}

// A Service implements modification logic for Songs.
type Service interface {
	// ParseArtists sets song.Artists based on other song fields.
	// This method should be used to process songs parsed from a TXT source where multiple artists are not supported.
	ParseArtists(ctx context.Context, song *model.Song)

	// Prepare prepares song for TXT serialization.
	// This includes merging multiple artists and setting appropriate file names.
	Prepare(ctx context.Context, song *model.Song)
}
