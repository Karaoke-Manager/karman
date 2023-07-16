package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/internal/api/v1"
	"github.com/Karaoke-Manager/karman/internal/service/upload"
	"github.com/Karaoke-Manager/karman/pkg/rwfs"
)

type Controller struct {
	v1Controller *v1.Controller
}

func NewController(db *gorm.DB, songFS rwfs.FS, uploadService upload.Service) *Controller {
	c := &Controller{
		v1Controller: v1.NewController(db, songFS, uploadService),
	}
	return c
}

func (c *Controller) Router(r chi.Router) {
	// Restrict requests to JSON for now
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.CleanPath)
	// TODO: Some CORS stuff
	// r.Use(middleware.Compress())
	// r.Use(middleware.RealIP)
	// r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Route("/v1", c.v1Controller.Router)

	r.NotFound(c.NotFound)
	r.MethodNotAllowed(c.MethodNotAllowed)
}
