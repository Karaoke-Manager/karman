package media

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/dbutil"
)

// dbRepo is a Repository implementation backed by a PostgreSQL database.
type dbRepo struct {
	db pgxutil.DB
}

// NewDBRepository returns a new Repository backed by the specified connection.
// db can be a single connection or a connection pool.
func NewDBRepository(db pgxutil.DB) Repository {
	return &dbRepo{db}
}

// CreateFile creates file in the database.
func (r *dbRepo) CreateFile(ctx context.Context, file *model.File) error {
	prepareFile(file)
	row, err := pgxutil.InsertRowReturning(ctx, r.db, "files", map[string]any{
		"type": file.Type,
	}, "uuid, created_at, updated_at", pgx.RowToStructByName[struct {
		UUID      uuid.UUID
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}])
	if err != nil {
		return err
	}
	file.UUID = row.UUID
	file.CreatedAt = row.CreatedAt
	file.UpdatedAt = row.UpdatedAt
	return nil
}

// UpdateFile updates file in the database.
func (r *dbRepo) UpdateFile(ctx context.Context, file *model.File) error {
	prepareFile(file)
	updatedAt, err := pgxutil.UpdateRowReturning(ctx, r.db, "files", map[string]any{
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
		return dbutil.Error(err)
	}
	file.UpdatedAt = updatedAt
	return nil
}

// prepareFile ensures that non-null fields are set to appropriate zero values.
func prepareFile(file *model.File) {
	if file.Checksum == nil {
		file.Checksum = make([]byte, 0)
	}
}
