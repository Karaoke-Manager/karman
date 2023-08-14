package v1

import (
	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/karman/api/v1/dav"
	"github.com/Karaoke-Manager/karman/api/v1/songs"
	"github.com/Karaoke-Manager/karman/api/v1/uploads"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
	"github.com/Karaoke-Manager/karman/service/upload"
)

// Controller implements the /v1 API namespace.
type Controller struct {
	uploadController *uploads.Controller
	songController   *songs.Controller
	davController    *dav.Controller
}

// NewController creates a new controller using the specified services.
// This function will create the required sub-controllers automatically.
func NewController(songService song.Service, mediaService media.Service, uploadService upload.Service) *Controller {
	return &Controller{
		uploadController: uploads.NewController(uploadService),
		songController:   songs.NewController(songService, mediaService),
		davController:    dav.NewController(songService, mediaService),
	}
}

// Router mounts the v1 sub-routers to r.
func (c *Controller) Router(r chi.Router) {
	r.Route("/songs", c.songController.Router)
	r.Route("/uploads", c.uploadController.Router)
	r.Route("/dav", c.davController.Router)
}
