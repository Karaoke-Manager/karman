package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/v1"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
	"github.com/Karaoke-Manager/karman/service/upload"
)

// Controller is the main API controller.
// This is basically the root entrypoint of the Karman API.
// All other API endpoints are created as sub-controllers of this controller.
type Controller struct {
	v1Controller *v1.Controller
}

// NewController creates a new Controller instance using the specified dependencies.
// The injected dependencies are passed along to the sub-controllers.
func NewController(songService song.Service, mediaService media.Service, uploadService upload.Service) *Controller {
	return &Controller{
		v1Controller: v1.NewController(songService, mediaService, uploadService),
	}
}

// Router sets up the router of this controller.
// This method is intended to be used to mount this controller as a sub-router of another chi.Router instance.
func (c *Controller) Router(r chi.Router) {
	// Restrict requests to JSON for now
	r.Use(middleware.CleanPath)
	// TODO: Some CORS stuff
	// r.Use(middleware.Compress())
	// r.Use(middleware.RealIP)
	// r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(render.NotAcceptableHandler(c.NotAcceptable))
	r.Route("/v1", c.v1Controller.Router)

	r.NotFound(c.NotFound)
	r.MethodNotAllowed(c.MethodNotAllowed)
}

// NotFound is an HTTP endpoint that renders a generic 404 Not Found error.
// This endpoint is the default 404 endpoint for the Controller and its sub-controllers.
func (*Controller) NotFound(w http.ResponseWriter, r *http.Request) {
	_ = render.Render(w, r, apierror.ErrNotFound)
}

// MethodNotAllowed is an HTTP endpoint that renders a generic 405 Method Not Allowed error.
// This endpoint is the default 405 endpoint for the Controller and its sub-controllers.
func (*Controller) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	// TODO: Include Allow Header: https://github.com/go-chi/chi/issues/446
	_ = render.Render(w, r, apierror.ErrMethodNotAllowed)
}

// NotAcceptable is an HTTP endpoint that renders a generic 406 Not Acceptable error.
// This endpoint is the default 406 endpoint fo the Controller and its sub-controllers.
func (*Controller) NotAcceptable(w http.ResponseWriter, r *http.Request) {
	_ = render.Render(w, r, apierror.ErrNotAcceptable)
}
