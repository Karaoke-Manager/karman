package songs

import (
	"github.com/Karaoke-Manager/go-ultrastar/txt"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"net/http"
)

func (c Controller) GetTxt(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	usSong := c.svc.SongData(song)

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
	c.svc.UpdateSongFromData(&song, usSong)
	err = c.svc.SaveSong(r.Context(), &song)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	s := schema.FromSong(song)
	_ = render.Render(w, r, &s)
}
