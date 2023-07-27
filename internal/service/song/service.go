package song

import (
	"context"

	"codello.dev/ultrastar"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/internal/model"
)

// Service provides an interface for working with songs.
type Service interface {
	// SaveSong persists the specified song.
	// This method may update fields of the song.
	// If an error occurs it is returned.
	SaveSong(ctx context.Context, song *model.Song) error

	// FindSongs retrieves a paginated view of songs from the persistence layer.
	// This method returns the page contents, the total number of songs and an error (if one occurred).
	FindSongs(ctx context.Context, limit, offset int) ([]model.Song, int64, error)

	// GetSong retrieves the song with the specified UUID from the persistence layer.
	// If an error occurs (such as the song not being found), the return value will indicate as much.
	GetSong(ctx context.Context, id uuid.UUID) (model.Song, error)

	// GetSongWithFiles fetches the song with the specified UUID from the persistence layer.
	// In contrast to GetSong this method will make sure that the song's media fields are populated correctly.
	GetSongWithFiles(ctx context.Context, id uuid.UUID) (model.Song, error)

	// DeleteSongByUUID deletes the song with the specified ID.
	// If no such song exists, this method does not return an error.
	DeleteSongByUUID(ctx context.Context, id uuid.UUID) error

	// SongData converts the specified model.Song into an equivalent ultrastar.Song.
	SongData(song model.Song) *ultrastar.Song

	// UpdateSongFromData updates song with the metadata from data.
	// The media fields of song are left untouched.
	UpdateSongFromData(song *model.Song, data *ultrastar.Song)
}

// NewService creates a new default implementation of Service using db as persistence layer.
func NewService(db *gorm.DB) Service {
	return service{db}
}

// service is the default Service implementation.
type service struct {
	db *gorm.DB
}
