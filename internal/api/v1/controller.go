package v1

import (
	"github.com/Karaoke-Manager/karman/internal/api/v1/songs"
	"github.com/Karaoke-Manager/karman/internal/api/v1/uploads"
	"github.com/Karaoke-Manager/karman/internal/service/media"
	"github.com/Karaoke-Manager/karman/internal/service/song"
	"github.com/Karaoke-Manager/karman/internal/service/upload"
	"github.com/go-chi/chi/v5"
)

// Controller implements the /v1 API namespace.
type Controller struct {
	uploadController *uploads.Controller
	songController   songs.Controller
}

// NewController creates a new controller using the specified services.
// This function will create the required sub-controllers automatically.
func NewController(songService song.Service, mediaService media.Service, uploadService upload.Service) Controller {
	return Controller{
		uploadController: uploads.NewController(uploadService),
		songController:   songs.NewController(songService, mediaService),
	}
}

// Router mounts the v1 sub-routers to r.
func (c Controller) Router(r chi.Router) {
	r.Route("/songs", c.songController.Router)
	r.Route("/uploads", c.uploadController.Router)
}
