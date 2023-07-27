package songs

import (
	"io"
	"net/http"
	"strconv"

	"codello.dev/ultrastar/txt"

	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// GetTxt implements the GET /v1/songs/{uuid}/txt endpoint.
func (c Controller) GetTxt(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	usSong := c.songSvc.SongData(song)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_ = txt.WriteSong(w, usSong)
}

// ReplaceTxt implements the PUT /v1/songs/{uuid}/txt endpoint.
func (c Controller) ReplaceTxt(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	usSong, err := txt.ReadSong(r.Body)
	if err != nil {
		_ = render.Render(w, r, apierror.InvalidUltraStarTXT(err))
		return
	}
	c.songSvc.UpdateSongFromData(&song, usSong)
	err = c.songSvc.SaveSong(r.Context(), &song)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	s := schema.FromSong(song)
	_ = render.Render(w, r, &s)
}

// GetCover implements the GET /v1/songs/{uuid}/cover endpoint.
func (c Controller) GetCover(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	if song.CoverFile == nil {
		_ = render.Render(w, r, apierror.MediaFileNotFound(song, "cover"))
		return
	}
	c.sendFile(w, r, *song.CoverFile)
}

// GetBackground implements the GET /v1/songs/{uuid}/background endpoint.
func (c Controller) GetBackground(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	if song.BackgroundFile == nil {
		_ = render.Render(w, r, apierror.MediaFileNotFound(song, "background"))
		return
	}
	c.sendFile(w, r, *song.BackgroundFile)
}

// GetAudio implements the GET /v1/songs/{uuid}/audio endpoint.
func (c Controller) GetAudio(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	if song.AudioFile == nil {
		_ = render.Render(w, r, apierror.MediaFileNotFound(song, "audio"))
		return
	}
	c.sendFile(w, r, *song.AudioFile)
}

// GetVideo implements the GET /v1/songs/{uuid}/video endpoint.
func (c Controller) GetVideo(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	if song.VideoFile == nil {
		_ = render.Render(w, r, apierror.MediaFileNotFound(song, "video"))
		return
	}
	c.sendFile(w, r, *song.VideoFile)
}

// sendFile sends the file as response to r.
// This method makes sure that the required headers are set.
func (c Controller) sendFile(w http.ResponseWriter, r *http.Request, file model.File) {
	f, err := c.mediaSvc.ReadFile(r.Context(), file)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", file.Type)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	w.WriteHeader(http.StatusOK)
	// TODO: Logging
	// The header is already written. We can't send error messages anymore
	_, _ = io.Copy(w, f)
}

// ReplaceCover implements the PUT /v1/songs/{uuid}/cover endpoint.
func (c Controller) ReplaceCover(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	mediaType := r.Header.Get("Content-Type")
	file, err := c.mediaSvc.StoreImageFile(r.Context(), mediaType, r.Body)
	if err != nil {
		// TODO: Logging
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	song.CoverFile = &file
	if err = c.songSvc.SaveSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// ReplaceBackground implements the PUT /v1/songs/{uuid}/background endpoint.
func (c Controller) ReplaceBackground(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	mediaType := r.Header.Get("Content-Type")
	file, err := c.mediaSvc.StoreImageFile(r.Context(), mediaType, r.Body)
	if err != nil {
		// TODO: Logging
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	song.BackgroundFile = &file
	if err = c.songSvc.SaveSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// DeleteCover implements the DELETE /v1/songs/{uuid}/cover endpoint.
func (c Controller) DeleteCover(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	song.CoverFileID = nil
	if err := c.songSvc.SaveSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// DeleteBackground implements the DELETE /v1/songs/{uuid}/background endpoint.
func (c Controller) DeleteBackground(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	song.BackgroundFileID = nil
	if err := c.songSvc.SaveSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// DeleteAudio implements the DELETE /v1/songs/{uuid}/audio endpoint.
func (c Controller) DeleteAudio(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	song.AudioFileID = nil
	if err := c.songSvc.SaveSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// DeleteVideo implements the DELETE /v1/songs/{uuid}/video endpoint.
func (c Controller) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	song.VideoFileID = nil
	if err := c.songSvc.SaveSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}
