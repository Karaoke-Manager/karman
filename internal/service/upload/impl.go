package upload

import (
	"context"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/pkg/rwfs"
	"gorm.io/gorm"
	"io"
	"io/fs"
)

type FS interface {
	rwfs.FS
	rwfs.MkDirFS
	rwfs.RemoveFS
}

type service struct {
	db *gorm.DB
	fs FS
}

func NewService(db *gorm.DB, fs FS) Service {
	return &service{db, fs}
}

func (s *service) CreateUpload(ctx context.Context) (upload model.Upload, err error) {
	db := s.db.WithContext(ctx)
	upload = model.NewUpload()
	if err = db.Create(&upload).Error; err != nil {
		return
	}
	// TODO: Maybe make the file mode configurable
	if err = s.fs.MkDir(upload.UUID.String(), fs.ModeDir&0o770); err != nil {
		// TODO: If an error occurs it should at least be logged.
		_ = db.Unscoped().Delete(&upload)
		return
	}
	return
}

func (s *service) GetUpload(ctx context.Context, uuid string) (upload model.Upload, err error) {
	err = s.db.WithContext(ctx).First(&upload, "uuid = ?", uuid).Error
	return
}

func (s *service) GetUploads(ctx context.Context, limit int, offset int) (uploads []model.Upload, err error) {
	err = s.db.WithContext(ctx).Find(&uploads).Limit(limit).Offset(offset).Error
	return
}

func (s *service) DeleteUploadByUUID(ctx context.Context, uuid string) (err error) {
	// TODO: Potentially stop processing
	err = s.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&model.Upload{}).Error
	// TODO: Delete files? Probably not on soft delete
	return
}

// TODO: Maybe use a chroot-style FS to prevent breakout.
func (s *service) CreateFile(ctx context.Context, upload model.Upload, path string, r io.Reader) error {
	if !upload.Open {
		return ErrUploadClosed
	}
	file, err := rwfs.Create(s.fs, upload.UUID.String()+"/"+path, 0660)
	if err != nil {
		return err
	}
	// TODO: Probably delete file if there is an error?
	// TODO: Handle Close error?
	defer file.Close()
	if _, err = io.Copy(file, r); err != nil {
		return err
	}
	return nil
}

func (s *service) StatFile(ctx context.Context, upload model.Upload, path string) (fs.FileInfo, error) {
	if !upload.Open {
		return nil, ErrUploadClosed
	}
	return fs.Stat(s.fs, upload.UUID.String()+"/"+path)
}

func (s *service) ReadDir(ctx context.Context, upload model.Upload, path string) ([]fs.DirEntry, error) {
	if !upload.Open {
		return nil, ErrUploadClosed
	}
	return fs.ReadDir(s.fs, upload.UUID.String()+"/"+path)
}

func (s *service) DeleteFile(ctx context.Context, upload model.Upload, path string) error {
	if !upload.Open {
		return ErrUploadClosed
	}
	return s.fs.Remove(upload.UUID.String() + "/" + path)
}

func (s *service) MarkForProcessing(ctx context.Context, upload model.Upload) error {
	if !upload.Open {
		return ErrUploadClosed
	}
	upload.Open = false
	upload.SongsTotal = -1
	upload.SongsProcessed = -1
	return s.db.WithContext(ctx).Save(&upload).Error
}

func (s *service) BeginProcessing(ctx context.Context, upload model.Upload) error {
	panic("not implemented")
}
