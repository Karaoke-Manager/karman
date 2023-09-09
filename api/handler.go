package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	v1 "github.com/Karaoke-Manager/karman/api/v1"
	"github.com/Karaoke-Manager/karman/core/media"
	"github.com/Karaoke-Manager/karman/core/song"
	"github.com/Karaoke-Manager/karman/core/upload"
	"github.com/Karaoke-Manager/karman/pkg/render"
	_ "github.com/Karaoke-Manager/karman/pkg/render/json" // JSON encoding for responses
)

// HealthChecker is an interface that can provide information about the system health.
type HealthChecker interface {
	// HealthCheck performs a health check and returns its result.
	// A result of true indicates that the system is healthy.
	HealthCheck(ctx context.Context) bool
}

// Handler is the main API handler.
// This is basically the root entrypoint of the Karman API.
// All other API endpoints are created as sub-handlers of this controller.
type Handler struct {
	r  chi.Router
	hc HealthChecker
}

// NewHandler creates a new Handler instance using the specified dependencies.
// The injected dependencies are passed along to the sub-handlers.
// debug indicates whether additional debugging features should be enabled.
func NewHandler(
	logger *slog.Logger,
	requestLogger *slog.Logger,
	hc HealthChecker,
	songRepo song.Repository,
	songSvc song.Service,
	mediaSvc media.Service,
	mediaStore media.Store,
	uploadRepo upload.Repository,
	uploadStore upload.Store,
	debug bool,
) *Handler {
	r := chi.NewRouter()
	h := &Handler{r, hc}
	v1Handler := v1.NewHandler(
		logger,
		songRepo,
		songSvc,
		mediaSvc,
		mediaStore,
		uploadRepo,
		uploadStore,
	)
	r.Use(middleware.Logger(requestLogger))
	r.Use(middleware.Recoverer(logger, debug))
	// Restrict requests to JSON for now
	r.Use(chimiddleware.CleanPath)
	// TODO: Some CORS stuff
	// TODO: Support running on subpath
	// r.Use(middleware.Compress())
	// r.Use(middleware.RealIP)
	r.Use(chimiddleware.StripSlashes)
	r.Use(render.NotAcceptableHandler(h.NotAcceptable))
	r.Mount("/v1", v1Handler)
	r.HandleFunc("/healthz", h.Healthz)

	r.NotFound(h.NotFound)
	r.MethodNotAllowed(h.MethodNotAllowed)
	return h
}

// ServeHTTP processes HTTP requests for h.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

// Healthz implements the /healthz endpoint.
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	if h.hc == nil || h.hc.HealthCheck(r.Context()) {
		_ = render.NoContent(w, r)
	} else {
		_ = render.Render(w, r, apierror.ErrServiceUnavailable)
	}
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
