package upload

import (
	"context"
	"errors"
	"github.com/Karaoke-Manager/karman/internal/entity"
	"io"
	"io/fs"
)

var (
	ErrUploadClosed = errors.New("upload closed")
)

type Service interface {
	CreateUpload(ctx context.Context) (entity.Upload, error)
	GetUpload(ctx context.Context, uuid string) (entity.Upload, error)
	FindUploads(ctx context.Context, limit int, offset int) ([]entity.Upload, int64, error)
	DeleteUploadByUUID(ctx context.Context, uuid string) error

	CreateFile(ctx context.Context, upload entity.Upload, path string, r io.Reader) error
	StatFile(ctx context.Context, upload entity.Upload, path string) (fs.FileInfo, error)
	ReadDir(ctx context.Context, upload entity.Upload, path string) ([]fs.DirEntry, error)
	DeleteFile(ctx context.Context, upload entity.Upload, path string) error
}
