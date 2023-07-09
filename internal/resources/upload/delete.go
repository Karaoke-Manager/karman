package upload

import (
	"github.com/Karaoke-Manager/karman/internal/apierror"
	"github.com/Karaoke-Manager/karman/internal/models"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (s *Server) Delete(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	var upload models.Upload
	if err := s.db.Where("uuid = ?", uuid).Delete(&upload).Error; err != nil {
		_ = apierror.DBError(w, r, err)
		return
	}
	_ = render.NoContent(w, r)
}
