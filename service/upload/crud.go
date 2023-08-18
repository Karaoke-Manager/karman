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
	return e.ToModel(0), nil
}

func (s *service) GetUpload(ctx context.Context, id uuid.UUID) (*model.Upload, error) {
	var e entity.Upload
	err := s.db.WithContext(ctx).First(&e, "uuid = ?", id).Error
	if err != nil {
		return nil, common.DBError(err)
	}

	var total int64
	if err = s.db.WithContext(ctx).Model(&entity.UploadProcessingError{}).Where("upload_id = ?", e.ID).Count(&total).Error; err != nil {
		return nil, common.DBError(err)
	}
	return e.ToModel(int(total)), nil
}

func (s *service) FindUploads(ctx context.Context, limit int, offset int64) ([]*model.Upload, int64, error) {
	var total int64
	var es []entity.Upload
	if err := s.db.WithContext(ctx).Model(&entity.Upload{}).Count(&total).Error; err != nil {
		return nil, total, common.DBError(err)
	}
	if err := s.db.WithContext(ctx).Find(&es).Limit(limit).Offset(int(offset)).Error; err != nil {
		return nil, total, common.DBError(err)
	}
	uploads := make([]*model.Upload, len(es))
	for i, e := range es {
		var errors int64
		if err := s.db.WithContext(ctx).Model(&entity.UploadProcessingError{}).Where("upload_id = ?", e.ID).Count(&errors).Error; err != nil {
			return nil, 0, common.DBError(err)
		}
		uploads[i] = e.ToModel(int(errors))
	}
	return uploads, total, nil
}

func (s *service) DeleteUpload(ctx context.Context, id uuid.UUID) error {
	// TODO: Stop processing
	err := s.db.WithContext(ctx).Where("uuid = ?", id).Delete(&entity.Upload{}).Error
	return common.DBError(err)
}
