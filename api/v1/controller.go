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
func NewController(
	songRepo song.Repository,
	songSvc song.Service,
	mediaSvc media.Service,
	mediaStore media.Store,
	uploadRepo upload.Repository,
	uploadStore upload.Store) *Controller {
	return &Controller{
		uploadController: uploads.NewController(uploadRepo, uploadStore),
		songController:   songs.NewController(songRepo, songSvc, mediaStore, mediaSvc),
		davController:    dav.NewController(songRepo, songSvc, mediaStore),
	}
}

// Router mounts the v1 sub-routers to r.
func (c *Controller) Router(r chi.Router) {
	r.Route("/songs", c.songController.Router)
	r.Route("/uploads", c.uploadController.Router)
	r.Route("/dav", c.davController.Router)
}
