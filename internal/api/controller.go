package api

import (
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/service/song"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"

	"github.com/Karaoke-Manager/karman/internal/api/v1"
	"github.com/Karaoke-Manager/karman/internal/service/upload"
)

type Controller struct {
	v1Controller *v1.Controller
}

func NewController(songService song.Service, uploadService upload.Service) *Controller {
	c := &Controller{
		v1Controller: v1.NewController(songService, uploadService),
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

func (c *Controller) NotFound(w http.ResponseWriter, r *http.Request) {
	_ = render.Render(w, r, apierror.ErrNotFound)
}

func (c *Controller) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	// TODO: Include Allow Header: https://github.com/go-chi/chi/issues/446
	_ = render.Render(w, r, apierror.ErrMethodNotAllowed)
}
