package upload

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
)

var (
	ErrUploadClosed = errors.New("upload closed")
)

type Service interface {
	CreateUpload(ctx context.Context) (*model.Upload, error)
	GetUpload(ctx context.Context, id uuid.UUID) (*model.Upload, error)
	FindUploads(ctx context.Context, limit int, offset int64) ([]*model.Upload, int64, error)
	DeleteUpload(ctx context.Context, id uuid.UUID) error

	/*CreateFile(ctx context.Context, upload *model.Upload, path string, r io.Reader) error
	StatFile(ctx context.Context, upload *model.Upload, path string) (fs.FileInfo, error)
	ReadDir(ctx context.Context, upload *model.Upload, path string) ([]fs.DirEntry, error)
	DeleteFile(ctx context.Context, upload *model.Upload, path string) error*/
}

type FS interface {
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db}
}
