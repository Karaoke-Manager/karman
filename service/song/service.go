package song

import (
	"context"
	"fmt"
	"mime"
	"strings"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// service is the default Service implementation.
type service struct{}

// NewService creates a new Service.
func NewService() Service {
	return &service{}
}

// ParseArtists splits song.Artist into song.Artists.
func (s *service) ParseArtists(ctx context.Context, song *model.Song) {
	artists := strings.Split(song.Artist, ",")
	for i, artist := range artists {
		artists[i] = strings.TrimSpace(artist)
	}
	// TODO: Parse featured artists as well
	song.Artists = artists
}

// Prepare sets song.Artist as well as file names for referenced files.
func (s *service) Prepare(ctx context.Context, song *model.Song) {
	song.Artist = strings.Join(song.Artists, ", ")
	// TODO: Generate safe file names
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
