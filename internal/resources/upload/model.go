package upload

import "github.com/Karaoke-Manager/karman/internal/models"

type Schema struct {
	// FIXME: can we use uuid.UUID type here?
	UUID   string              `json:"id"`
	Status models.UploadStatus `json:"status"`
}

func SchemaFromModel(m models.Upload) Schema {
	return Schema{
		UUID:   m.UUID.String(),
		Status: m.Status,
	}
}
