package song

import (
	"context"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/karman/internal/model"
)

func (s service) CreateSong(ctx context.Context, data *ultrastar.Song) (model.Song, error) {
	song := model.NewSongWithData(data)
	err := s.db.WithContext(ctx).Save(&song).Error
	return song, err
}

func (s service) FindSongs(ctx context.Context, limit, offset int) (songs []model.Song, total int64, err error) {
	if err = s.db.WithContext(ctx).Model(&model.Song{}).Count(&total).Error; err != nil {
		return
	}
	if err = s.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&songs).Error; err != nil {
		return
	}
	return
}

func (s service) GetSong(ctx context.Context, uuid string) (song model.Song, err error) {
	err = s.db.WithContext(ctx).
		First(&song, "uuid = ?", uuid).Error
	return
}

func (s service) GetSongWithFiles(ctx context.Context, uuid string) (song model.Song, err error) {
	err = s.db.WithContext(ctx).
		Joins("AudioFile").
		Joins("VideoFile").
		Joins("CoverFile").
		Joins("BackgroundFile").
		First(&song, "songs.uuid = ?", uuid).Error
	return
}

func (s service) SaveSong(ctx context.Context, song *model.Song) error {
	return s.db.WithContext(ctx).Save(song).Error
}

func (s service) DeleteSongByUUID(ctx context.Context, uuid string) error {
	return s.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&model.Song{}).Error
}
