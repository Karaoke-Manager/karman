package media

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

func (db *db) CreateFile(ctx context.Context, mediaType mediatype.MediaType) (*model.File, error) {
	row, err := pgxutil.InsertRowReturning(ctx, db.q, "files", map[string]any{
		"type": mediaType,
	}, "uuid, created_at, updated_at", pgx.RowToStructByName[struct {
		UUID      uuid.UUID
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}])
	if err != nil {
		return nil, err
	}

	return &model.File{
		Model: model.Model{
			UUID:      row.UUID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Type: mediaType,
	}, nil
}

func (db *db) UpdateFile(ctx context.Context, file *model.File) error {
	updatedAt, err := pgxutil.UpdateRowReturning(ctx, db.q, "files", map[string]any{
		"type":     file.Type,
		"size":     file.Size,
		"checksum": file.Checksum,
		"duration": file.Duration,
		"width":    file.Width,
		"height":   file.Height,
	}, map[string]any{
		"uuid": file.UUID,
	}, "updated_at", pgx.RowTo[time.Time])
	if err != nil {
		return err
	}
	file.UpdatedAt = updatedAt
	return nil
}
