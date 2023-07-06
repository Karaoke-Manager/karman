package upload

import (
	"github.com/Karaoke-Manager/karman/internal/apierror"
	"github.com/Karaoke-Manager/karman/internal/models"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"net/http"
)

type CreateRequestSchema struct {
}

type CreateResponseSchema struct {
	Schema
}

func (s *Server) Create(w http.ResponseWriter, r *http.Request) {
	var data CreateRequestSchema
	if err := render.Decode(r, &data); err != nil {
		_ = apierror.DecodeError(w, r, err)
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

	resp := CreateResponseSchema{SchemaFromModel(upload)}
	_ = render.Respond(w, r, resp)
}
