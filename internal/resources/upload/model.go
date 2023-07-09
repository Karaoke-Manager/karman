package upload

import (
	"github.com/Karaoke-Manager/karman/internal/models"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

type Schema struct {
	render.NopRenderer
	// FIXME: should we use uuid.UUID type here?
	UUID   string              `json:"id"`
	Status models.UploadStatus `json:"status"`
}

func (s *Server) SchemaFromModel(m models.Upload) Schema {
	return Schema{
		UUID:   m.UUID.String(),
		Status: m.Status,
	}
}
