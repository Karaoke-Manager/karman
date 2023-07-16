package songs

import (
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	var song model.Song
	if err := c.db.First(&song, "uuid = ?", uuid).Error; err != nil {
		_ = render.Render(w, r, apierror.DBError(err))
		return
	}

	_ = render.Render(w, r, resp)
}
