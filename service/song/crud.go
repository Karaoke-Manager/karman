package song

import (
	"context"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/entity"
)

// FindSongs fetches a page of songs from the database.
func (s *service) FindSongs(ctx context.Context, limit, offset int) ([]*model.Song, int64, error) {
	var total int64
	var es []entity.Song
	if err := s.db.WithContext(ctx).Model(&entity.Song{}).Where("upload_id IS NULL").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := s.db.WithContext(ctx).Model(&entity.Song{}).
		Joins("AudioFile").
		Joins("VideoFile").
		Joins("CoverFile").
		Joins("BackgroundFile").
		Where("songs.upload_id IS NULL").Limit(limit).Offset(offset).
		Find(&es).Error; err != nil {
		return nil, total, err
	}
	songs := make([]*model.Song, len(es))
	for i, e := range es {
		songs[i] = e.ToModel()
		s.ensureFilenames(songs[i])
	}
	return songs, total, nil
}

// GetSong fetches a single song from the database.
func (s *service) GetSong(ctx context.Context, id uuid.UUID) (*model.Song, error) {
	var e entity.Song
	if err := s.db.WithContext(ctx).
		Joins("AudioFile").
		Joins("VideoFile").
		Joins("CoverFile").
		Joins("BackgroundFile").
		First(&e, "songs.uuid = ?", id).Error; err != nil {
		return nil, err
	}
	song := e.ToModel()
	s.ensureFilenames(song)
	return song, nil
}

// CreateSong persists a new song into the database.
// If song already exists in the database, an error is returned.
func (s *service) CreateSong(ctx context.Context, song *model.Song) error {
	e := entity.SongFromModel(song)
	err := s.db.WithContext(ctx).Create(&e).Error
	song.UUID = e.UUID
	song.CreatedAt = e.CreatedAt
	song.UpdatedAt = e.UpdatedAt
	return err
}

// UpdateSongData updates song in the database.
// song must already have been persisted before.
func (s *service) UpdateSongData(ctx context.Context, song *model.Song) error {
	e := entity.SongFromModel(song)
	return s.db.WithContext(ctx).Model(&e).
		Where("uuid = ?", song.UUID).
		Select("*").Omit(
		"ID", "UUID",
		"CreatedAt",
		"DeletedAt",
		"UploadID",
		"AudioFileID",
		"VideoFileID",
		"CoverFileID",
		"BackgroundFileID").
		Updates(&e).Error
}

// DeleteSong deletes the song with the specified UUID from the database.
func (s *service) DeleteSong(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Where("uuid = ?", id).Delete(&entity.Song{}).Error
}
