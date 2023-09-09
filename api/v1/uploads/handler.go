package uploads

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/Karaoke-Manager/karman/service/upload"
)

// Handler implements the /v1/uploads endpoints.
type Handler struct {
	r      chi.Router
	logger *slog.Logger

	uploadRepo  upload.Repository
	uploadStore upload.Store
}

// NewHandler creates a new Handler instance using the specified service.
func NewHandler(logger *slog.Logger, uploadRepo upload.Repository, uploadStore upload.Store) *Handler {
	r := chi.NewRouter()
	h := &Handler{r, logger.With("log", "uploads.handler"), uploadRepo, uploadStore}

	r.With(render.ContentTypeNegotiation("application/json")).Post("/", h.Create)
	r.With(middleware.Paginate(25, 100), render.ContentTypeNegotiation("application/json")).Get("/", h.Find)

	r.Group(func(r chi.Router) {
		r.Use(middleware.UUID("uuid"))
		r.Delete("/{uuid}", h.Delete)

		r.Group(func(r chi.Router) {
			r.Use(h.FetchUpload)
			r.With(render.ContentTypeNegotiation("application/json")).Get("/{uuid}", h.Get)

			r.Group(func(r chi.Router) {
				r.Use(ValidateFilePath, UploadState(model.UploadStateOpen))
				r.With(middleware.RequireContentType("application/octet-stream")).Put("/{uuid}/files/*", h.PutFile)
				r.With(render.ContentTypeNegotiation("application/json")).Get("/{uuid}/files/*", h.GetFile)
				r.With(render.ContentTypeNegotiation("application/json")).Get("/{uuid}/files", h.GetFile)
				r.Delete("/{uuid}/files/*", h.DeleteFile)
				// the following routes always return an error but are included for API consistency
				r.With(middleware.RequireContentType("application/octet-stream")).Put("/{uuid}/files", h.PutFile)
				r.Delete("/{uuid}/files", h.DeleteFile)
			})

			r.Group(func(r chi.Router) {
				r.Use(UploadState(model.UploadStateProcessing, model.UploadStateDone))
				r.With(middleware.Paginate(100, 1000), render.ContentTypeNegotiation("application/json")).Get("/{uuid}/errors", h.GetErrors)
			})
		})
	})

	// POST /{uuid}/beginProcessing

	// GET /{uuid}/songs

	// OPTION 1:
	// GET /{uuid}/songs/{id2}
	// DELETE /{uuid}/songs{id2}
	// POST /{uuid}/import

	// OPTION 2:
	// GET /{uuid}/songs
	// GET /songs/...
	// DELETE /songs/...
	// POST /{uuid}/import (with the songs to import)
	return h
}

// ServeHTTP processes HTTP requests for h.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}
