package song

import (
	"context"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/internal/model"
)

// FindSongs fetches a page of songs from the database.
func (s service) FindSongs(ctx context.Context, limit, offset int) (songs []model.Song, total int64, err error) {
	if err = s.db.WithContext(ctx).Model(&model.Song{}).Where("upload_id IS NULL").Count(&total).Error; err != nil {
		return
	}
	if err = s.db.WithContext(ctx).Model(&model.Song{}).Where("upload_id IS NULL").Limit(limit).Offset(offset).Find(&songs).Error; err != nil {
		return
	}
	return
}

// GetSong fetches a single song from the database.
func (s service) GetSong(ctx context.Context, id uuid.UUID) (song model.Song, err error) {
	err = s.db.WithContext(ctx).
		First(&song, "uuid = ?", id).Error
	return
}

// GetSongWithFiles fetches a single song with its file references from the database.
func (s service) GetSongWithFiles(ctx context.Context, id uuid.UUID) (song model.Song, err error) {
	err = s.db.WithContext(ctx).
		Joins("AudioFile").
		Joins("VideoFile").
		Joins("CoverFile").
		Joins("BackgroundFile").
		First(&song, "songs.uuid = ?", id).Error
	return
}

// SaveSong persists song into the database.
func (s service) SaveSong(ctx context.Context, song *model.Song) error {
	return s.db.WithContext(ctx).Save(song).Error
}

// DeleteSongByUUID deletes the song with the specified UUID from the database.
func (s service) DeleteSongByUUID(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Where("uuid = ?", id).Delete(&model.Song{}).Error
}
