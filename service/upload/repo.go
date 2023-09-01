package upload

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/dbutil"
)

// dbRepo is the main Repository implementation, backed by a PostgreSQL database.
type dbRepo struct {
	db pgxutil.DB
}

// NewDBRepository creates a new Repository backed by the specified database connection.
// db can be a single connection or a connection pool.
func NewDBRepository(db pgxutil.DB) Repository {
	return &dbRepo{db}
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
		return model.Upload{}, dbutil.Error(err)
	}
	return row.toModel(), nil
}

// FindUploads lists uploads with pagination.
func (r *dbRepo) FindUploads(ctx context.Context, limit int, offset int64) ([]model.Upload, int64, error) {
	total, err := pgxutil.SelectRow(ctx, r.db, `SELECT COUNT(*) FROM uploads`, nil, pgx.RowTo[int64])
	if err != nil {
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
		return nil, total, err
	}
	return uploads, total, nil
}

// DeleteUpload deletes the upload with the specified UUID.
func (r *dbRepo) DeleteUpload(ctx context.Context, id uuid.UUID) (bool, error) {
	// TODO: Stop processing
	_, err := pgxutil.ExecRow(ctx, r.db, `DELETE FROM uploads WHERE uuid = $1`, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
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
		return nil, 0, dbutil.Error(err)
	}
	uploadErrors, err := pgxutil.Select(ctx, r.db, `SELECT
    file, message
	FROM upload_errors
	WHERE upload_id = $1
	LIMIT CASE WHEN $2 < 0 THEN NULL ELSE $2 END OFFSET $3`, []any{row.ID, limit, offset}, pgx.RowToStructByName[model.UploadProcessingError])
	if err != nil {
		return nil, row.Total, err
	}
	return uploadErrors, row.Total, nil
}
