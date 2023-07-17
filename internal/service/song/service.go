package song

import (
	"context"
	"github.com/Karaoke-Manager/karman/internal/model"
)

type Service interface {
	CreateSong(ctx context.Context) (model.Song, error)
	GetSong(ctx context.Context, uuid string) (model.Song, error)
}
