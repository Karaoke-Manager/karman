package upload

import (
	"context"
	"errors"
	"github.com/Karaoke-Manager/karman/internal/model"
	"io"
	"io/fs"
)

var (
	ErrUploadClosed = errors.New("upload closed")
)

type Service interface {
	CreateUpload(ctx context.Context) (model.Upload, error)
	GetUpload(ctx context.Context, uuid string) (model.Upload, error)
	GetUploads(ctx context.Context, limit int, offset int) ([]model.Upload, error)
	DeleteUploadByUUID(ctx context.Context, uuid string) error

	CreateFile(ctx context.Context, upload model.Upload, path string, r io.Reader) error
	StatFile(ctx context.Context, upload model.Upload, path string) (fs.FileInfo, error)
	ReadDir(ctx context.Context, upload model.Upload, path string) ([]fs.DirEntry, error)
	DeleteFile(ctx context.Context, upload model.Upload, path string) error
}
