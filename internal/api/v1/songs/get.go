package songs

import (
	"github.com/Karaoke-Manager/karman/pkg/render"
	"net/http"
)

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	_ = render.Render(w, r, resp)
}
