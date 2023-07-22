package songs

import (
	"github.com/Karaoke-Manager/go-ultrastar/txt"
	"net/http"
)

func (c *Controller) GetTxt(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	usSong := c.svc.UltraStarSong(r.Context(), song)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_ = txt.WriteSong(w, usSong)
}
