package upload

import (
	"context"
	"errors"
	"io"
	"io/fs"

	"github.com/Karaoke-Manager/karman/model"
)

var (
	ErrUploadClosed = errors.New("upload closed")
)

type Service interface {
	CreateUpload(ctx context.Context) (*model.Upload, error)
	GetUpload(ctx context.Context, uuid string) (*model.Upload, error)
	FindUploads(ctx context.Context, limit int, offset int64) ([]*model.Upload, int64, error)
	DeleteUploadByUUID(ctx context.Context, uuid string) error

	CreateFile(ctx context.Context, upload *model.Upload, path string, r io.Reader) error
	StatFile(ctx context.Context, upload *model.Upload, path string) (fs.FileInfo, error)
	ReadDir(ctx context.Context, upload *model.Upload, path string) ([]fs.DirEntry, error)
	DeleteFile(ctx context.Context, upload *model.Upload, path string) error
}