package songs

import (
	"io"
	"mime"
	"net/http"
	"strconv"

	"codello.dev/ultrastar/txt"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// GetTxt implements the GET /v1/songs/{uuid}/txt endpoint.
func (c *Controller) GetTxt(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	c.songSvc.Prepare(r.Context(), &song)

	t := render.MustGetNegotiatedContentType(r)
	if t.Equals(mediatype.TextPlain) {
		t = t.WithoutParameters("charset", "utf-8")
	}
	w.Header().Set("Content-Type", t.String())
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": song.TxtFileName}))
	w.WriteHeader(http.StatusOK)
	_ = txt.WriteSong(w, song.Song)
}

// ReplaceTxt implements the PUT /v1/songs/{uuid}/txt endpoint.
func (c *Controller) ReplaceTxt(w http.ResponseWriter, r *http.Request) {
	var err error
	song := MustGetSong(r.Context())
	song.Song, err = txt.NewReader(r.Body).ReadSong()
	if err != nil {
		_ = render.Render(w, r, apierror.InvalidUltraStarTXT(err))
		return
	}
	c.songSvc.ParseArtists(r.Context(), &song)
	err = c.songRepo.UpdateSong(r.Context(), &song)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	c.songSvc.Prepare(r.Context(), &song)
	s := schema.FromSong(song)
	_ = render.Render(w, r, &s)
}

// GetCover implements the GET /v1/songs/{uuid}/cover endpoint.
func (c *Controller) GetCover(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	if song.CoverFile == nil {
		_ = render.Render(w, r, apierror.MediaFileNotFound(song, "cover"))
		return
	}
	c.songSvc.Prepare(r.Context(), &song)
	c.sendFile(w, r, song.CoverFile, song.CoverFileName)
}

// GetBackground implements the GET /v1/songs/{uuid}/background endpoint.
func (c *Controller) GetBackground(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	if song.BackgroundFile == nil {
		_ = render.Render(w, r, apierror.MediaFileNotFound(song, "background"))
		return
	}
	c.songSvc.Prepare(r.Context(), &song)
	c.sendFile(w, r, song.BackgroundFile, song.BackgroundFileName)
}

// GetAudio implements the GET /v1/songs/{uuid}/audio endpoint.
func (c *Controller) GetAudio(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	if song.AudioFile == nil {
		_ = render.Render(w, r, apierror.MediaFileNotFound(song, "audio"))
		return
	}
	c.songSvc.Prepare(r.Context(), &song)
	c.sendFile(w, r, song.AudioFile, song.AudioFileName)
}

// GetVideo implements the GET /v1/songs/{uuid}/video endpoint.
func (c *Controller) GetVideo(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	if song.VideoFile == nil {
		_ = render.Render(w, r, apierror.MediaFileNotFound(song, "video"))
		return
	}
	c.songSvc.Prepare(r.Context(), &song)
	c.sendFile(w, r, song.VideoFile, song.VideoFileName)
}

// sendFile sends the file as response to r.
// This method makes sure that the required headers are set.
func (c *Controller) sendFile(w http.ResponseWriter, r *http.Request, file *model.File, name string) {
	contentType := render.NegotiateContentType(r, file.Type)
	if contentType.IsNil() {
		render.NotAcceptable(w, r)
		return
	}
	f, err := c.mediaStore.Open(r.Context(), file.Type, file.UUID)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	w.Header().Set("Content-Type", contentType.String())
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": name}))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}

// ReplaceCover implements the PUT /v1/songs/{uuid}/cover endpoint.
func (c *Controller) ReplaceCover(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	mediaType := mediatype.MustParse(r.Header.Get("Content-Type"))
	file, err := c.mediaSvc.StoreFile(r.Context(), mediaType, r.Body)
	if err != nil {
		// TODO: Logging
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	song.CoverFile = &file
	if err = c.songRepo.UpdateSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// ReplaceBackground implements the PUT /v1/songs/{uuid}/background endpoint.
func (c *Controller) ReplaceBackground(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	mediaType := mediatype.MustParse(r.Header.Get("Content-Type"))
	file, err := c.mediaSvc.StoreFile(r.Context(), mediaType, r.Body)
	if err != nil {
		// TODO: Logging
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	song.BackgroundFile = &file
	if err = c.songRepo.UpdateSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// ReplaceAudio implements the PUT /v1/songs/{uuid}/audio endpoint.
func (c *Controller) ReplaceAudio(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	mediaType := mediatype.MustParse(r.Header.Get("Content-Type"))
	file, err := c.mediaSvc.StoreFile(r.Context(), mediaType, r.Body)
	if err != nil {
		// TODO: Logging
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	song.AudioFile = &file
	if err = c.songRepo.UpdateSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// ReplaceVideo implements the PUT /v1/songs/{uuid}/video endpoint.
func (c *Controller) ReplaceVideo(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	mediaType := mediatype.MustParse(r.Header.Get("Content-Type"))
	file, err := c.mediaSvc.StoreFile(r.Context(), mediaType, r.Body)
	if err != nil {
		// TODO: Logging
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	song.VideoFile = &file
	if err = c.songRepo.UpdateSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// DeleteCover implements the DELETE /v1/songs/{uuid}/cover endpoint.
func (c *Controller) DeleteCover(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	song.CoverFile = nil
	if err := c.songRepo.UpdateSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// DeleteBackground implements the DELETE /v1/songs/{uuid}/background endpoint.
func (c *Controller) DeleteBackground(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	song.BackgroundFile = nil
	if err := c.songRepo.UpdateSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// DeleteAudio implements the DELETE /v1/songs/{uuid}/audio endpoint.
func (c *Controller) DeleteAudio(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	song.AudioFile = nil
	if err := c.songRepo.UpdateSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// DeleteVideo implements the DELETE /v1/songs/{uuid}/video endpoint.
func (c *Controller) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	song.VideoFile = nil
	if err := c.songRepo.UpdateSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}
