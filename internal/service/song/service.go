package song

import (
	"context"

	"github.com/Karaoke-Manager/server/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service provides an interface for working with songs.
type Service interface {
	CreateSong(ctx context.Context, song *model.Song) error

	// UpdateSongData updates the data of an existing song in the persistence layer (metadata and music).
	// If an error occurs it is returned.
	UpdateSongData(ctx context.Context, song *model.Song) error

	// FindSongs retrieves a paginated view of songs from the persistence layer.
	// This method returns the page contents, the total number of songs and an error (if one occurred).
	FindSongs(ctx context.Context, limit, offset int) ([]*model.Song, int64, error)

	// GetSong retrieves the song with the specified UUID from the persistence layer.
	// If an error occurs (such as the song not being found), the return value will indicate as much.
	GetSong(ctx context.Context, id uuid.UUID) (*model.Song, error)

	// DeleteSong deletes the song with the specified ID.
	// If no such song exists, this method does not return an error.
	DeleteSong(ctx context.Context, id uuid.UUID) error

	// ReplaceCover sets the cover of song to file and persists the change.
	// Both the song and file must already be persisted.
	ReplaceCover(ctx context.Context, song *model.Song, file *model.File) error

	// ReplaceAudio sets the audio of song to file and persists the change.
	// Both the song and file must already be persisted.
	ReplaceAudio(ctx context.Context, song *model.Song, file *model.File) error

	// ReplaceVideo sets the video of song to file and persists the change.
	// Both the song and file must already be persisted.
	ReplaceVideo(ctx context.Context, song *model.Song, file *model.File) error

	// ReplaceBackground sets the background of song to file and persists the change.
	// Both the song and file must already be persisted.
	ReplaceBackground(ctx context.Context, song *model.Song, file *model.File) error
}

// NewService creates a new default implementation of Service using db as persistence layer.
func NewService(db *gorm.DB) Service {
	return &service{db}
}

// service is the default Service implementation.
type service struct {
	db *gorm.DB
}
