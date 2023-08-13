package schema

import (
	"github.com/Karaoke-Manager/karman/internal/entity"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

type UploadStatus string

const (
	UploadStatusCreated    UploadStatus = "created"
	UploadStatusPending    UploadStatus = "pending"
	UploadStatusProcessing UploadStatus = "processing"
	UploadStatusReady      UploadStatus = "ready"
)

type Upload struct {
	render.NopRenderer
	UUID   string       `json:"id"`
	Status UploadStatus `json:"status"`
}

func NewUploadFromModel(m entity.Upload) *Upload {
	var status UploadStatus
	if m.Open {
		status = UploadStatusCreated
	} else if m.SongsProcessed == -1 {
		status = UploadStatusPending
	} else if m.SongsProcessed != m.SongsTotal {
		status = UploadStatusProcessing
	} else {
		status = UploadStatusReady
	}
	return &Upload{
		UUID:   m.UUID.String(),
		Status: status,
	}
}
