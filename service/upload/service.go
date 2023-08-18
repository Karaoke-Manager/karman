package upload

import (
	"context"
	"io"
	"io/fs"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
)

type Service interface {
	CreateUpload(ctx context.Context) (*model.Upload, error)
	GetUpload(ctx context.Context, id uuid.UUID) (*model.Upload, error)
	FindUploads(ctx context.Context, limit int, offset int64) ([]*model.Upload, int64, error)
	DeleteUpload(ctx context.Context, id uuid.UUID) error

	CreateFile(ctx context.Context, upload *model.Upload, path string) (io.WriteCloser, error)
	StatFile(ctx context.Context, upload *model.Upload, path string) (fs.FileInfo, error)
	OpenDir(ctx context.Context, upload *model.Upload, path string) (Dir, error)
	DeleteFile(ctx context.Context, upload *model.Upload, path string) error

	GetErrors(ctx context.Context, upload *model.Upload, limit int, offset int64) ([]*model.UploadProcessingError, int64, error)
}

func NewService(db *gorm.DB, store Store) Service {
	return &service{db, store}
}

type service struct {
	db    *gorm.DB
	store Store
}
