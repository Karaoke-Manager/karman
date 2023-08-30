package media

import (
	"context"

	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

type DB interface {
	CreateFile(ctx context.Context, mediaType mediatype.MediaType) (*model.File, error)
	UpdateFile(ctx context.Context, file *model.File) error
}

type db struct {
	q pgxutil.DB
}

func NewDB(pool pgxutil.DB) DB {
	return &db{pool}
}
