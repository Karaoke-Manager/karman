package song

import (
	"context"
	"fmt"
	"mime"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/service/common"
	"github.com/Karaoke-Manager/karman/service/entity"
)

// ReplaceCover sets song.CoverFile to file and persists the change in the database.
func (s *service) ReplaceCover(ctx context.Context, song *model.Song, file *model.File) error {
	var id any
	if file != nil {
		id = s.db.Model(&entity.File{}).Select("ID").Where("uuid = ?", file.UUID)
	}
	err := s.db.WithContext(ctx).Model(&entity.Song{}).Where("uuid = ?", song.UUID).Update("CoverFileID", id).Error
	if err == nil {
		song.CoverFile = file
	}
	return common.DBError(err)
}

// ReplaceAudio sets song.AudioFile to file and persists the change in the database.
func (s *service) ReplaceAudio(ctx context.Context, song *model.Song, file *model.File) error {
	var id any
	if file != nil {
		id = s.db.Model(&entity.File{}).Select("ID").Where("uuid = ?", file.UUID)
	}
	err := s.db.WithContext(ctx).Model(&entity.Song{}).Where("uuid = ?", song.UUID).Update("AudioFileID", id).Error
	if err == nil {
		song.AudioFile = file
	}
	return common.DBError(err)
}

// ReplaceVideo sets song.VideoFile to file and persists the change in the database.
func (s *service) ReplaceVideo(ctx context.Context, song *model.Song, file *model.File) error {
	var id any
	if file != nil {
		id = s.db.Model(&entity.File{}).Select("ID").Where("uuid = ?", file.UUID)
	}
	err := s.db.WithContext(ctx).Model(&entity.Song{}).Where("uuid = ?", song.UUID).Update("VideoFileID", id).Error
	if err == nil {
		song.VideoFile = file
	}
	return common.DBError(err)
}

// ReplaceBackground sets song.BackgroundFile to file and persists the change in the database.
func (s *service) ReplaceBackground(ctx context.Context, song *model.Song, file *model.File) error {
	var id any
	if file != nil {
		id = s.db.Model(&entity.File{}).Select("ID").Where("uuid = ?", file.UUID)
	}
	err := s.db.WithContext(ctx).Model(&entity.Song{}).Where("uuid = ?", song.UUID).Update("BackgroundFileID", id).Error
	if err == nil {
		song.BackgroundFile = file
	}
	return common.DBError(err)
}

// ensureFilenames sets the different file name fields of song.Song.
// If song does not have a file, the respective field is not modified.
func (s *service) ensureFilenames(song *model.Song) {
	song.TxtFileName = fmt.Sprintf("%s - %s.txt", song.Artist, song.Title)
	if song.AudioFile != nil {
		song.AudioFileName = fmt.Sprintf("%s - %s [AUDIO]%s", song.Artist, song.Title, s.extensionForType(song.AudioFile.Type))
	}
	if song.CoverFile != nil {
		song.CoverFileName = fmt.Sprintf("%s - %s [CO]%s", song.Artist, song.Title, s.extensionForType(song.CoverFile.Type))
	}
	if song.VideoFile != nil {
		song.VideoFileName = fmt.Sprintf("%s - %s [VIDEO]%s", song.Artist, song.Title, s.extensionForType(song.VideoFile.Type))
	}
	if song.BackgroundFile != nil {
		song.BackgroundFileName = fmt.Sprintf("%s - %s [BG]%s", song.Artist, song.Title, s.extensionForType(song.BackgroundFile.Type))
	}
}

// extensionForType returns the file extension that should be used for the specified media type.
// The returned extension includes a leading dot.
func (*service) extensionForType(t mediatype.MediaType) string {
	// preferred, known types
	switch t.FullType() {
	case "audio/mpeg", "audio/mp3":
		return ".mp3"
	case "video/mp4":
		return ".mp4"
	case "image/jpeg":
		return ".jpg"
	}
	ext, _ := mime.ExtensionsByType(t.FullType())
	if len(ext) == 0 {
		return ""
	}
	return ext[0]
}
