package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Karaoke-Manager/karman/api/apierror"
	v1 "github.com/Karaoke-Manager/karman/api/v1"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
	"github.com/Karaoke-Manager/karman/service/upload"
)

// Handler is the main API handler.
// This is basically the root entrypoint of the Karman API.
// All other API endpoints are created as sub-handlers of this controller.
type Handler struct {
	r chi.Router
}

// NewHandler creates a new Handler instance using the specified dependencies.
// The injected dependencies are passed along to the sub-handlers.
func NewHandler(songRepo song.Repository, songSvc song.Service, mediaSvc media.Service, mediaStore media.Store, uploadRepo upload.Repository, uploadStore upload.Store) *Handler {
	r := chi.NewRouter()
	h := &Handler{r}
	v1Handler := v1.NewHandler(songRepo, songSvc, mediaSvc, mediaStore, uploadRepo, uploadStore)
	// Restrict requests to JSON for now
	r.Use(middleware.CleanPath)
	// TODO: Some CORS stuff
	// r.Use(middleware.Compress())
	// r.Use(middleware.RealIP)
	// r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(render.NotAcceptableHandler(h.NotAcceptable))
	r.Mount("/v1", v1Handler)

	r.NotFound(h.NotFound)
	r.MethodNotAllowed(h.MethodNotAllowed)
	return h
}

// ServeHTTP processes HTTP requests for h.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

// NotFound is an HTTP endpoint that renders a generic 404 Not Found error.
// This endpoint is the default 404 endpoint for the Handler and its sub-handlers.
func (*Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	_ = render.Render(w, r, apierror.ErrNotFound)
}

// MethodNotAllowed is an HTTP endpoint that renders a generic 405 Method Not Allowed error.
// This endpoint is the default 405 endpoint for the Handler and its sub-handlers.
func (*Handler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	// TODO: Include Allow Header: https://github.com/go-chi/chi/issues/446
	_ = render.Render(w, r, apierror.ErrMethodNotAllowed)
}

// NotAcceptable is an HTTP endpoint that renders a generic 406 Not Acceptable error.
// This endpoint is the default 406 endpoint fo the Handler and its sub-handlers.
func (*Handler) NotAcceptable(w http.ResponseWriter, r *http.Request) {
	_ = render.Render(w, r, apierror.ErrNotAcceptable)
}
