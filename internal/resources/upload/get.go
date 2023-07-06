package upload

import (
	"github.com/Karaoke-Manager/karman/internal/apierror"
	"github.com/Karaoke-Manager/karman/internal/models"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type GetResponseSchema struct {
	Schema
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	var upload models.Upload
	if err := s.db.First(&upload, "uuid = ?", uuid).Error; err != nil {
		_ = apierror.DBError(w, r, err)
		return
	}
	resp := GetResponseSchema{SchemaFromModel(upload)}
	_ = render.Respond(w, r, resp)
}
