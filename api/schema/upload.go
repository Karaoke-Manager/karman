package schema

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

type Upload struct {
	render.NopRenderer
	UUID   uuid.UUID         `json:"id"`
	Status model.UploadState `json:"status"`

	SongsTotal     int `json:"songsTotal"`
	SongsProcessed int `json:"songsProcessed"`
	Errors         int `json:"errors"`
}

func FromUpload(m *model.Upload) Upload {
	return Upload{
		UUID:           m.UUID,
		Status:         m.State,
		SongsTotal:     m.SongsTotal,
		SongsProcessed: m.SongsProcessed,
		Errors:         m.Errors,
	}
}

func (u *Upload) PrepareResponse(w http.ResponseWriter, r *http.Request) any {
	switch u.Status {
	case model.UploadStateOpen, model.UploadStatePending:
		return map[string]any{
			"uuid":   u.UUID,
			"status": u.Status,
		}
	case model.UploadStateProcessing:
		return u
	case model.UploadStateDone:
		return map[string]any{
			"uuid":       u.UUID,
			"status":     u.Status,
			"songsTotal": u.SongsTotal,
			"errors":     u.Errors,
		}
	}
	return u
}
