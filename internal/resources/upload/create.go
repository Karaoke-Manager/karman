package upload

import (
	"github.com/Karaoke-Manager/karman/internal/apierror"
	"github.com/Karaoke-Manager/karman/internal/models"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"net/http"
)

type CreateRequestSchema struct {
	render.NopBinder
}

type CreateResponseSchema struct {
	Schema
}

func (s *Server) Create(w http.ResponseWriter, r *http.Request) {
	var data CreateRequestSchema
	if err := render.Bind(r, &data); err != nil {
		_ = apierror.BindError(w, r, err)
		return
	}

	// Create the upload
	upload := models.Upload{
		Status: models.UploadStatusCreated,
	}
	if err := s.db.Create(&upload).Error; err != nil {
		// FIXME: Maybe check for validation errors here?
		_ = apierror.DBError(w, r, err)
		return
	}

	resp := CreateResponseSchema{s.SchemaFromModel(upload)}
	_ = render.Render(w, r, resp)
}
