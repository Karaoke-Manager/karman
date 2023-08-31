package upload

import (
	"context"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/common"
	"github.com/Karaoke-Manager/karman/service/entity"
)

// GetErrors lists processing errors for an upload with pagination.
func (s *service) GetErrors(ctx context.Context, upload *model.Upload, limit int, offset int64) ([]*model.UploadProcessingError, int64, error) {
	var e entity.Upload
	if err := s.db.WithContext(ctx).Where("uuid = ?", upload.UUID).First(&e).Error; err != nil {
		return nil, 0, common.DBError(err)
	}

	var total int64
	var errors []entity.UploadProcessingError
	if err := s.db.WithContext(ctx).Model(&entity.UploadProcessingError{}).Where("upload_id = ?", e.ID).Count(&total).Error; err != nil {
		return nil, 0, common.DBError(err)
	}
	if err := s.db.WithContext(ctx).Where("upload_id = ?", e.ID).Limit(limit).Offset(int(offset)).Find(&errors).Error; err != nil {
		return nil, total, common.DBError(err)
	}

	errorModels := make([]*model.UploadProcessingError, len(errors))
	for i, err := range errors {
		errorModels[i] = err.ToModel()
	}
	return errorModels, total, nil
}
