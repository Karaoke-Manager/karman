package media

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgxutil"
	"github.com/lmittmann/tint"

	"github.com/Karaoke-Manager/karman/core/internal/dbutil"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// dbRepo is a Repository implementation backed by a PostgreSQL database.
type dbRepo struct {
	logger *slog.Logger
	db     pgxutil.DB
}

// NewDBRepository returns a new Repository backed by the specified connection.
// db can be a single connection or a connection pool.
func NewDBRepository(logger *slog.Logger, db pgxutil.DB) Repository {
	return &dbRepo{logger, db}
}

// fileRow is the data returned by a SELECT query for files.
// This type is used by Get and Find operations.
type fileRow struct {
	UUID      uuid.UUID
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
	DeletedAt pgtype.Timestamp `db:"deleted_at"`

	Path string

	Type     mediatype.MediaType
	Size     int64
	Checksum []byte
	Duration time.Duration
	Width    int
	Height   int
}

// toModel converts r into an equivalent model.Song.
func (r fileRow) toModel() model.File {
	f := model.File{
		Model: model.Model{
			UUID:      r.UUID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			DeletedAt: time.Time{},
		},
		UploadPath: r.Path,
		Type:       r.Type,
		Size:       r.Size,
		Checksum:   r.Checksum,
		Duration:   r.Duration,
		Width:      r.Width,
		Height:     r.Height,
	}
	if r.DeletedAt.Valid {
		f.DeletedAt = r.DeletedAt.Time
	}
	return f
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
		r.logger.ErrorContext(ctx, "Could not create media file.", "type", file.Type, tint.Err(err))
		return err
	}
	file.UUID = row.UUID
	file.CreatedAt = row.CreatedAt
	file.UpdatedAt = row.UpdatedAt
	return nil
}

// GetFile fetches a file from the database.
func (r *dbRepo) GetFile(ctx context.Context, id uuid.UUID) (model.File, error) {
	row, err := pgxutil.SelectRow(ctx, r.db, `SELECT
    uuid, created_at, updated_at, deleted_at,
    CASE WHEN upload_id IS NULL THEN '' ELSE path END AS path,
    type, size, checksum, duration, width, height
    FROM files
    WHERE uuid = $1`, []any{id}, pgx.RowToStructByName[fileRow])
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			r.logger.ErrorContext(ctx, "Could not fetch file.", "uuid", id, tint.Err(err))
		}
		return model.File{}, dbutil.Error(err)
	}
	return row.toModel(), nil
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
		if !errors.Is(err, pgx.ErrNoRows) {
			r.logger.ErrorContext(ctx, "Could not update file.", "uuid", file.UUID, tint.Err(err))
		}
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

// DeleteFile immediately deletes the file with the specified ID from the database.
func (r *dbRepo) DeleteFile(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := pgxutil.ExecRow(ctx, r.db, `DELETE FROM files WHERE uuid = $1`, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not delete file.", "uuid", id, tint.Err(err))
		return false, err
	}
	return true, nil
}

// FindOrphanedFiles returns a list of files that do not belong to an upload or a song.
func (r *dbRepo) FindOrphanedFiles(ctx context.Context, limit int64) ([]model.File, error) {
	files, err := pgxutil.Select(ctx, r.db, `SELECT DISTINCT
    uuid, created_at, updated_at, deleted_at,
    CASE WHEN upload_id IS NULL THEN '' ELSE path END AS path,
    type, size, checksum, duration, width, height
    FROM files f
	WHERE upload_id IS NULL AND NOT EXISTS(
	    SELECT s.id FROM songs s WHERE s.audio_file_id = f.id OR s.cover_file_id = f.id OR s.video_file_id = f.id OR s.background_file_id = f.id
	)
	LIMIT CASE WHEN $1 < 0 THEN NULL ELSE $1 END`, []any{limit}, func(row pgx.CollectableRow) (model.File, error) {
		data, err := pgx.RowToStructByName[fileRow](row)
		return data.toModel(), err
	})
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not list orphaned files.", "limit", limit, tint.Err(err))
	}
	return files, err
}
