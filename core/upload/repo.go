package upload

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
)

// dbRepo is the main Repository implementation, backed by a PostgreSQL database.
type dbRepo struct {
	logger *slog.Logger
	db     pgxutil.DB
}

// NewDBRepository creates a new Repository backed by the specified database connection.
// db can be a single connection or a connection pool.
func NewDBRepository(logger *slog.Logger, db pgxutil.DB) Repository {
	return &dbRepo{logger, db}
}

// uploadRow is the data returned by a SELECT query for uploads.
type uploadRow struct {
	UUID           uuid.UUID
	CreatedAt      time.Time        `db:"created_at"`
	UpdatedAt      time.Time        `db:"updated_at"`
	DeletedAt      pgtype.Timestamp `db:"deleted_at"`
	Open           bool
	SongsTotal     int `db:"songs_total"`
	SongsProcessed int `db:"songs_processed"`
	Errors         int
}

// toModel converts r to an equivalent model.Upload.
func (r uploadRow) toModel() model.Upload {
	upload := model.Upload{
		Model: model.Model{
			UUID:      r.UUID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		SongsTotal:     r.SongsTotal,
		SongsProcessed: r.SongsProcessed,
		Errors:         r.Errors,
	}
	if r.DeletedAt.Valid {
		upload.DeletedAt = r.DeletedAt.Time
	}
	switch {
	case r.Open:
		upload.State = model.UploadStateOpen
	case r.SongsTotal < 0:
		upload.State = model.UploadStatePending
	case r.SongsProcessed < r.SongsTotal:
		upload.State = model.UploadStateProcessing
	default:
		upload.State = model.UploadStateDone
	}
	return upload
}

// CreateUpload creates a new upload in the database.
func (r *dbRepo) CreateUpload(ctx context.Context, upload *model.Upload) error {
	row, err := pgxutil.InsertRowReturning(ctx, r.db, "uploads", map[string]any{
		"open": true,
	}, "uuid, created_at, updated_at, deleted_at, open, songs_total, songs_processed, 0 AS errors", pgx.RowToStructByName[uploadRow])
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not create upload.", tint.Err(err))
		return err
	}
	*upload = row.toModel()
	return nil
}

// GetUpload fetches an upload from the database.
func (r *dbRepo) GetUpload(ctx context.Context, id uuid.UUID) (model.Upload, error) {
	row, err := pgxutil.SelectRow(ctx, r.db, `SELECT
    uuid, created_at, updated_at, deleted_at,
    open, songs_total, songs_processed, COUNT(upload_errors.id) AS errors
	FROM uploads
	LEFT OUTER JOIN upload_errors ON upload_id = uploads.id
	WHERE uuid = $1
	GROUP BY uploads.id`, []any{id}, pgx.RowToStructByName[uploadRow])
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			r.logger.ErrorContext(ctx, "Could not fetch upload.", tint.Err(err))
		}
		return model.Upload{}, dbutil.Error(err)
	}
	return row.toModel(), nil
}

// FindUploads lists uploads with pagination.
func (r *dbRepo) FindUploads(ctx context.Context, limit int, offset int64) ([]model.Upload, int64, error) {
	total, err := pgxutil.SelectRow(ctx, r.db, `SELECT COUNT(*) FROM uploads`, nil, pgx.RowTo[int64])
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not count uploads.", "limit", limit, "offset", offset, tint.Err(err))
		return nil, 0, err
	}
	uploads, err := pgxutil.Select(ctx, r.db, `SELECT
    uuid, created_at, updated_at, deleted_at,
    open, songs_total, songs_processed, COUNT(upload_errors.id) AS errors
	FROM uploads
	LEFT OUTER JOIN upload_errors ON upload_id = uploads.id
	GROUP BY uploads.id
	LIMIT CASE WHEN $1 < 0 THEN NULL ELSE $1 END OFFSET $2`, []any{limit, offset}, func(row pgx.CollectableRow) (model.Upload, error) {
		data, err := pgx.RowToStructByName[uploadRow](row)
		return data.toModel(), err
	})
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not list uploads.", "limit", limit, "offset", offset, tint.Err(err))
		return nil, total, err
	}
	return uploads, total, nil
}

// UpdateUpload updates the upload in the database with upload.UUID.
func (r *dbRepo) UpdateUpload(ctx context.Context, upload *model.Upload) error {
	updatedAt, err := pgxutil.UpdateRowReturning(ctx, r.db, "uploads", map[string]any{
		"open":            upload.State == model.UploadStateOpen,
		"songs_total":     upload.SongsTotal,
		"songs_processed": upload.SongsProcessed,
	}, map[string]any{
		"uuid": upload.UUID,
	}, "updated_at", pgx.RowTo[time.Time])
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			r.logger.ErrorContext(ctx, "Could not update upload.", "uuid", upload.UUID, tint.Err(err))
		}
		return dbutil.Error(err)
	}
	upload.UpdatedAt = updatedAt
	return nil
}

// DeleteUpload deletes the upload with the specified UUID.
func (r *dbRepo) DeleteUpload(ctx context.Context, id uuid.UUID) (bool, error) {
	// TODO: Stop processing
	_, err := pgxutil.ExecRow(ctx, r.db, `DELETE FROM uploads WHERE uuid = $1`, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not delete upload.", tint.Err(err))
		return false, err
	}
	return true, nil
}

// CreateError creates a processing error for an upload.
func (r *dbRepo) CreateError(ctx context.Context, upload *model.Upload, processingError model.UploadProcessingError) error {
	_, err := pgxutil.ExecRow(ctx, r.db, `INSERT INTO upload_errors (upload_id, file, message)
		VALUES ((SELECT uploads.id FROM uploads WHERE uuid = $1), $2, $3)`,
		upload.UUID,
		processingError.File,
		processingError.Message,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not create upload error.", "uuid", upload.UUID, tint.Err(err))
		return err
	}
	upload.Errors++
	return nil
}

// GetErrors lists processing errors for an upload with pagination.
func (r *dbRepo) GetErrors(ctx context.Context, id uuid.UUID, limit int, offset int64) ([]model.UploadProcessingError, int64, error) {
	row, err := pgxutil.SelectRow(ctx, r.db, `SELECT
    uploads.id, COUNT(upload_errors.id)
	FROM uploads
	JOIN upload_errors ON upload_id = uploads.id
	WHERE uuid = $1
	GROUP BY uploads.id`, []any{id}, pgx.RowToStructByPos[struct {
		ID    int
		Total int64
	}])
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			r.logger.ErrorContext(ctx, "Could not count upload errors.", "uuid", id, "limit", limit, "offset", offset, tint.Err(err))
		}
		return nil, 0, dbutil.Error(err)
	}
	uploadErrors, err := pgxutil.Select(ctx, r.db, `SELECT
    file, message
	FROM upload_errors
	WHERE upload_id = $1
	LIMIT CASE WHEN $2 < 0 THEN NULL ELSE $2 END OFFSET $3`, []any{row.ID, limit, offset}, pgx.RowToStructByName[model.UploadProcessingError])
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not list upload errors.", "uuid", id, "limit", limit, "offset", offset, tint.Err(err))
		return nil, row.Total, err
	}
	return uploadErrors, row.Total, nil
}

// ClearErrors deletes all errors associated with the specified upload.
func (r *dbRepo) ClearErrors(ctx context.Context, upload *model.Upload) (bool, error) {
	_, err := pgxutil.ExecRow(ctx, r.db, `DELETE FROM upload_errors WHERE upload_id = $1`, upload.UUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not delete upload errors.", "uuid", upload.UUID, tint.Err(err))
		return false, err
	}
	upload.Errors = 0
	return true, nil
}

// ClearSongs deletes all songs associated with the specified upload.
func (r *dbRepo) ClearSongs(ctx context.Context, upload *model.Upload) (bool, error) {
	_, err := pgxutil.ExecRow(ctx, r.db, `DELETE FROM songs WHERE upload_id = $1`, upload.UUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not delete upload songs.", "uuid", upload.UUID, tint.Err(err))
	}
	upload.SongsTotal = -1
	upload.SongsProcessed = 0
	return true, nil
}

// ClearFiles deletes all files associated with the specified upload.
func (r *dbRepo) ClearFiles(ctx context.Context, upload *model.Upload) (bool, error) {
	_, err := pgxutil.ExecRow(ctx, r.db, `DELETE FROM files WHERE upload_id = $1`, upload.UUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		r.logger.ErrorContext(ctx, "Could not delete upload files.", "uuid", upload.UUID, tint.Err(err))
	}
	return true, nil
}
