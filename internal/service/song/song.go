package song

import (
	"fmt"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/karman/internal/model"
	"mime"
)

func (s service) UpdateSongFromData(song *model.Song, data *ultrastar.Song) {
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
}

func (s service) SongData(song model.Song) *ultrastar.Song {
	// TODO: Make sure that relations are loaded
	customTags := make(map[string]string, len(song.Extra))
	for key, value := range song.Extra {
		customTags[key] = value
	}
	usSong := &ultrastar.Song{
		Gap:             song.Gap,
		VideoGap:        song.VideoGap,
		NotesGap:        song.NotesGap,
		Start:           song.Start,
		End:             song.End,
		PreviewStart:    song.PreviewStart,
		MedleyStartBeat: song.MedleyStartBeat,
		MedleyEndBeat:   song.MedleyEndBeat,
		CalcMedley:      song.CalcMedley,
		Resolution:      4,
		Title:           song.Title,
		Artist:          song.Artist,
		Genre:           song.Genre,
		Edition:         song.Edition,
		Creator:         song.Creator,
		Language:        song.Language,
		Year:            song.Year,
		Comment:         song.Comment,
		DuetSinger1:     song.DuetSinger1,
		DuetSinger2:     song.DuetSinger2,
		CustomTags:      customTags,
		MusicP1:         song.MusicP1.Clone(),
		MusicP2:         song.MusicP2.Clone(),
	}
	// TODO: Disambiguation if multiple files have the same extension
	if song.AudioFileID != nil {
		usSong.AudioFile = s.preferredAudioName(song)
	}
	if song.VideoFileID != nil {
		usSong.VideoFile = s.preferredVideoName(song)
	}
	if song.CoverFileID != nil {
		usSong.CoverFile = s.preferredCoverName(song)
	}
	if song.BackgroundFileID != nil {
		usSong.BackgroundFile = s.preferredBackgroundName(song)
	}
	return usSong
}

func (s service) preferredAudioName(song model.Song) string {
	return fmt.Sprintf("%s - %s [AUDIO]%s", song.Artist, song.Title, s.extensionForType(song.AudioFile.Type))
}

func (s service) preferredVideoName(song model.Song) string {
	return fmt.Sprintf("%s - %s [VIDEO]%s", song.Artist, song.Title, s.extensionForType(song.VideoFile.Type))
}

func (s service) preferredCoverName(song model.Song) string {
	return fmt.Sprintf("%s - %s [CO]%s", song.Artist, song.Title, s.extensionForType(song.CoverFile.Type))
}

func (s service) preferredBackgroundName(song model.Song) string {
	return fmt.Sprintf("%s - %s [BG]%s", song.Artist, song.Title, s.extensionForType(song.BackgroundFile.Type))
}

func (service) extensionForType(t string) string {
	if t == "audio/mpeg" {
		// special case for MP3 as this media type encompasses MP2 as well.
		return ".mp3"
	}
	ext, _ := mime.ExtensionsByType(t)
	if len(ext) == 0 {
		return ""
	}
	return ext[0]
}
