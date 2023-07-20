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

// Controller is the main API controller.
// This is basically the root entrypoint of the Karman API.
// All other API endpoints are created as sub-controllers of this controller.
type Controller struct {
	v1Controller *v1.Controller
}

// NewController creates a new Controller instance using the specified dependencies.
// The injected dependencies are passed along to the sub-controllers.
func NewController(songService song.Service, uploadService upload.Service) Controller {
	c := Controller{
		v1Controller: v1.NewController(songService, uploadService),
	}
	return c
}

// Router sets up the router of this controller.
// This method is intended to be used to mount this controller as a sub-router of another chi.Router instance.
func (c Controller) Router(r chi.Router) {
	// Restrict requests to JSON for now
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

// NotFound is an HTTP endpoint that renders a generic 404 Not Found error.
// This endpoint is the default 404 endpoint for the Controller and its sub-controllers.
func (c Controller) NotFound(w http.ResponseWriter, r *http.Request) {
	_ = render.Render(w, r, apierror.ErrNotFound)
}

// MethodNotAllowed is an HTTP endpoint that renders a generic 405 Method Not Allowed error.
// This endpoint is the default 405 endpoint for the Controller and its sub-controllers.
func (c Controller) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	// TODO: Include Allow Header: https://github.com/go-chi/chi/issues/446
	_ = render.Render(w, r, apierror.ErrMethodNotAllowed)
}
