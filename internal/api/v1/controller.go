package v1

import (
	"github.com/Karaoke-Manager/karman/internal/api/v1/songs"
	"github.com/Karaoke-Manager/karman/internal/api/v1/uploads"
	"github.com/Karaoke-Manager/karman/internal/service/upload"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Controller struct {
	uploadController *uploads.Controller
	songController   *songs.Controller
}

func NewController(db *gorm.DB, uploadService upload.Service) *Controller {
	s := &Controller{
		uploadController: uploads.NewController(uploadService),
		songController:   songs.NewController(db, songFS),
	}
	return s
}

func (c *Controller) Router(r chi.Router) {
	r.Route("/songs", c.songController.Router)
	r.Route("/uploads", c.uploadController.Router)
}
