package song

import (
	"context"

	"github.com/Karaoke-Manager/karman/model"

	"github.com/google/uuid"
)

// Service provides an interface for working with songs.
type Service interface {
	CreateSong(ctx context.Context, song *model.Song) error

	// SaveSong updates the data of an existing song in the persistence layer (metadata and music).
	// If an error occurs it is returned.
	SaveSong(ctx context.Context, song *model.Song) error

	// FindSongs retrieves a paginated view of songs from the persistence layer.
	// If limit = -1, all songs are returned.
	// This method returns the page contents, the total number of songs and an error (if one occurred).
	FindSongs(ctx context.Context, limit int, offset int64) ([]*model.Song, int64, error)

	// GetSong retrieves the song with the specified UUID from the persistence layer.
	// If an error occurs (such as the song not being found), the return value will indicate as much.
	GetSong(ctx context.Context, id uuid.UUID) (*model.Song, error)

	// DeleteSong deletes the song with the specified ID.
	// If no such song exists, this method does not return an error.
	DeleteSong(ctx context.Context, id uuid.UUID) error
}

// NewService creates a new default implementation of Service using db as persistence layer.
func NewService(db DB) Service {
	return &service{db}
}

// service is the default Service implementation.
type service struct {
	db DB
}
