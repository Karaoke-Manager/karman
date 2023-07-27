package song

import (
	"context"

	"codello.dev/ultrastar"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/internal/model"
)

type Service interface {
	SaveSong(ctx context.Context, song *model.Song) error
	FindSongs(ctx context.Context, limit, offset int) ([]model.Song, int64, error)
	GetSong(ctx context.Context, id uuid.UUID) (model.Song, error)
	GetSongWithFiles(ctx context.Context, id uuid.UUID) (model.Song, error)
	DeleteSongByUUID(ctx context.Context, id uuid.UUID) error

	SongData(song model.Song) *ultrastar.Song
	UpdateSongFromData(song *model.Song, data *ultrastar.Song)
}

func NewService(db *gorm.DB) Service {
	return service{db}
}

type service struct {
	db *gorm.DB
}
