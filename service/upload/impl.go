package upload

import (
	"context"
	"io"
	"io/fs"

	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/entity"
)

type FS interface {
}

type service struct {
	db *gorm.DB
	fs FS
}

func NewService(db *gorm.DB, fs FS) Service {
	return &service{db, fs}
}

func (s *service) CreateUpload(ctx context.Context) (*model.Upload, error) {
	db := s.db.WithContext(ctx)
	e := entity.Upload{}
	if err := db.Create(&e).Error; err != nil {
		return nil, err
	}
	return e.ToModel(), nil
}

func (s *service) GetUpload(ctx context.Context, uuid string) (upload *model.Upload, err error) {
	err = s.db.WithContext(ctx).First(&upload, "uuid = ?", uuid).Error
	return
}

func (s *service) FindUploads(ctx context.Context, limit int, offset int64) (uploads []*model.Upload, total int64, err error) {
	if err = s.db.WithContext(ctx).Find(&uploads).Count(&total).Error; err != nil {
		return
	}
	if err = s.db.WithContext(ctx).Find(&uploads).Limit(limit).Offset(int(offset)).Error; err != nil {
		return
	}
	return
}

func (s *service) DeleteUploadByUUID(ctx context.Context, uuid string) error {
	// TODO: Potentially stop processing
	return s.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&entity.Upload{}).Error
	// TODO: Delete files? Probably not on soft delete
}

// TODO: Maybe use a chroot-style FS to prevent breakout.
func (s *service) CreateFile(ctx context.Context, upload *model.Upload, path string, r io.Reader) error {
	if !upload.Open {
		return ErrUploadClosed
	}
	/*file, err := rwfs.Create(s.fs, upload.UUID.String()+"/"+path, 0660)
	if err != nil {
		return err
	}
	// TODO: Probably delete file if there is an error?
	// TODO: Handle Close error?
	defer file.Close()
	if _, err = io.Copy(file, r); err != nil {
		return err
	}*/
	return nil
}

func (s *service) StatFile(ctx context.Context, upload *model.Upload, path string) (fs.FileInfo, error) {
	if !upload.Open {
		return nil, ErrUploadClosed
	}
	//return fs.Stat(s.fs, upload.UUID.String()+"/"+path)
	return nil, nil
}

func (s *service) ReadDir(ctx context.Context, upload *model.Upload, path string) ([]fs.DirEntry, error) {
	if !upload.Open {
		return nil, ErrUploadClosed
	}
	// return fs.ReadDir(s.fs, upload.UUID.String()+"/"+path)
	return nil, nil
}

func (s *service) DeleteFile(ctx context.Context, upload *model.Upload, path string) error {
	if !upload.Open {
		return ErrUploadClosed
	}
	// return s.fs.Remove(upload.UUID.String() + "/" + path)
	return nil
}

func (s *service) MarkForProcessing(ctx context.Context, upload *model.Upload) error {
	if !upload.Open {
		return ErrUploadClosed
	}
	upload.Open = false
	upload.SongsTotal = -1
	upload.SongsProcessed = -1
	return s.db.WithContext(ctx).Save(&upload).Error
}

func (s *service) BeginProcessing(ctx context.Context, upload *model.Upload) error {
	panic("not implemented")
}
