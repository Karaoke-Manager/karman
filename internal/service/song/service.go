package song

import (
	"context"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/karman/internal/model"
	"gorm.io/gorm"
)

type Service interface {
	FindSongs(ctx context.Context, limit, offset int) ([]model.Song, int64, error)
	CreateSong(ctx context.Context, data *ultrastar.Song) (model.Song, error)
	GetSong(ctx context.Context, uuid string) (model.Song, error)
	GetSongWithFiles(ctx context.Context, uuid string) (model.Song, error)
	SaveSong(ctx context.Context, song *model.Song) error
	DeleteSongByUUID(ctx context.Context, uuid string) error
	UltraStarSong(ctx context.Context, song model.Song) *ultrastar.Song
}

func NewService(db *gorm.DB) Service {
	return service{db}
}

type service struct {
	db *gorm.DB
}
