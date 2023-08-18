package upload

import (
	"context"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/common"
	"github.com/Karaoke-Manager/karman/service/entity"
)

func (s *service) CreateUpload(ctx context.Context) (*model.Upload, error) {
	db := s.db.WithContext(ctx)
	e := entity.Upload{
		Open:           true,
		SongsTotal:     -1,
		SongsProcessed: -1,
	}
	if err := db.Create(&e).Error; err != nil {
		return nil, common.DBError(err)
	}
	return e.ToModel(), nil
}

func (s *service) GetUpload(ctx context.Context, id uuid.UUID) (*model.Upload, error) {
	var e entity.Upload
	err := s.db.WithContext(ctx).Preload("ProcessingErrors").First(&e, "uuid = ?", id).Error
	if err != nil {
		return nil, common.DBError(err)
	}
	return e.ToModel(), nil
}

func (s *service) FindUploads(ctx context.Context, limit int, offset int64) ([]*model.Upload, int64, error) {
	var total int64
	var es []entity.Upload
	if err := s.db.WithContext(ctx).Model(&entity.Upload{}).Count(&total).Error; err != nil {
		return nil, total, common.DBError(err)
	}
	if err := s.db.WithContext(ctx).Preload("ProcessingErrors").Find(&es).Limit(limit).Offset(int(offset)).Error; err != nil {
		return nil, total, common.DBError(err)
	}
	uploads := make([]*model.Upload, len(es))
	for i, e := range es {
		uploads[i] = e.ToModel()
	}
	return uploads, total, nil
}

func (s *service) DeleteUpload(ctx context.Context, id uuid.UUID) error {
	// TODO: Stop processing
	err := s.db.WithContext(ctx).Where("uuid = ?", id).Delete(&entity.Upload{}).Error
	return common.DBError(err)
}

/*
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
/*return nil
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
*/
