package songs

import (
	"github.com/Karaoke-Manager/go-ultrastar/txt"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"io"
	"net/http"
	"strconv"
)

func (c Controller) GetTxt(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	usSong := c.songSvc.SongData(song)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_ = txt.WriteSong(w, usSong)
}

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

func (c Controller) GetCover(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	if song.CoverFile == nil {
		_ = render.Render(w, r, apierror.MediaFileNotFound(song, "cover"))
		return
	}
	file, err := c.mediaSvc.ReadFile(r.Context(), *song.CoverFile)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", song.CoverFile.Type)
	w.Header().Set("Content-Length", strconv.FormatInt(song.CoverFile.Size, 10))
	w.WriteHeader(http.StatusOK)
	// The header is already written. We can't send error messages anymore
	_, _ = io.Copy(w, file)
}

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

func (c Controller) DeleteCover(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	song.CoverFileID = nil
	if err := c.songSvc.SaveSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}
