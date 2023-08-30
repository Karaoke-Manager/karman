package song

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
)

type DB interface {
	CreateSong(ctx context.Context, song *model.Song) error
	GetSong(ctx context.Context, id uuid.UUID) (*model.Song, error)
	FindSongs(ctx context.Context, limit int, offset int64) ([]*model.Song, int64, error)
	DeleteSongByUUID(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateSong(ctx context.Context, song *model.Song) error
}

type db struct {
	q pgxutil.DB
}

func NewDB(pool pgxutil.DB) DB {
	return &db{pool}
}
