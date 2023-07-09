package upload

import (
	"github.com/Karaoke-Manager/karman/internal/apierror"
	"github.com/Karaoke-Manager/karman/internal/models"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"net/http"
)

func (s *Server) List(w http.ResponseWriter, r *http.Request) {
	// TODO: Probably pagination
	var uploads []models.Upload
	if err := s.db.Find(&uploads).Error; err != nil {
		_ = apierror.DBError(w, r, err)
		return
	}
	uploadSchemas := make([]Schema, len(uploads))
	for i, upload := range uploads {
		uploadSchemas[i] = s.SchemaFromModel(upload)
	}
	resp := schema.List[Schema]{
		Items: uploadSchemas,
	}
	_ = render.Render(w, r, resp)
}
