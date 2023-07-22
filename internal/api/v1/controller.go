package v1

import (
	"github.com/Karaoke-Manager/karman/internal/api/v1/songs"
	"github.com/Karaoke-Manager/karman/internal/api/v1/uploads"
	"github.com/Karaoke-Manager/karman/internal/service/song"
	"github.com/Karaoke-Manager/karman/internal/service/upload"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	uploadController *uploads.Controller
	songController   songs.Controller
}

func NewController(songService song.Service, uploadService upload.Service) Controller {
	return Controller{
		uploadController: uploads.NewController(uploadService),
		songController:   songs.NewController(songService),
	}
}

func (c Controller) Router(r chi.Router) {
	r.Route("/songs", c.songController.Router)
	r.Route("/uploads", c.uploadController.Router)
}
