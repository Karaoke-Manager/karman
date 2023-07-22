package song

import (
	"context"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/google/uuid"
)

func (s service) CreateSong(ctx context.Context, data *ultrastar.Song) (model.Song, error) {
	song := model.NewSongWithData(data)
	err := s.db.WithContext(ctx).Save(&song).Error
	return song, err
}

func (s service) ReplaceSong(ctx context.Context, song *model.Song, data *ultrastar.Song) error {
	song.Gap = data.Gap
	song.VideoGap = data.VideoGap
	song.NotesGap = data.NotesGap
	song.Start = data.Start
	song.End = data.End
	song.PreviewStart = data.PreviewStart
	song.MedleyStartBeat = data.MedleyStartBeat
	song.MedleyEndBeat = data.MedleyEndBeat
	song.CalcMedley = data.CalcMedley
	song.Title = data.Title
	song.Artist = data.Artist
	song.Genre = data.Genre
	song.Edition = data.Edition
	song.Creator = data.Creator
	song.Language = data.Language
	song.Year = data.Year
	song.Comment = data.Comment
	song.DuetSinger1 = data.DuetSinger1
	song.DuetSinger2 = data.DuetSinger2
	song.Extra = data.CustomTags
	song.MusicP1 = data.MusicP1.Clone()
	song.MusicP2 = data.MusicP2.Clone()
	return s.db.WithContext(ctx).Save(&song).Error
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

func (s service) GetSong(ctx context.Context, id uuid.UUID) (song model.Song, err error) {
	err = s.db.WithContext(ctx).
		First(&song, "uuid = ?", id).Error
	return
}

func (s service) GetSongWithFiles(ctx context.Context, id uuid.UUID) (song model.Song, err error) {
	err = s.db.WithContext(ctx).
		Joins("AudioFile").
		Joins("VideoFile").
		Joins("CoverFile").
		Joins("BackgroundFile").
		First(&song, "songs.uuid = ?", id).Error
	return
}

func (s service) SaveSong(ctx context.Context, song *model.Song) error {
	return s.db.WithContext(ctx).Save(song).Error
}

func (s service) DeleteSongByUUID(ctx context.Context, uuid string) error {
	return s.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&model.Song{}).Error
}
