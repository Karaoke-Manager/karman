package uploads

import (
	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/Karaoke-Manager/karman/service/upload"
)

// Controller implements the /v1/uploads endpoints.
type Controller struct {
	svc upload.Service
}

// NewController creates a new Controller instance using the specified service.
func NewController(svc upload.Service) *Controller {
	s := &Controller{svc}
	return s
}

// Router sets up the routing for the endpoint.
func (c *Controller) Router(r chi.Router) {
	r.With(render.ContentTypeNegotiation("application/json")).Post("/", c.Create)
	r.With(middleware.Paginate(25, 100), render.ContentTypeNegotiation("application/json")).Get("/", c.Find)

	r.Group(func(r chi.Router) {
		r.Use(middleware.UUID("uuid"))
		r.Delete("/{uuid}", c.Delete)

		r.Group(func(r chi.Router) {
			r.Use(c.FetchUpload)
			r.With(render.ContentTypeNegotiation("application/json")).Get("/{uuid}", c.Get)

			r.Group(func(r chi.Router) {
				r.Use(ValidateFilePath, UploadState(model.UploadStateOpen))
				r.With(middleware.RequireContentType("application/octet-stream")).Put("/{uuid}/files/*", c.PutFile)
				r.With(render.ContentTypeNegotiation("application/json")).Get("/{uuid}/files/*", c.GetFile)
				r.With(render.ContentTypeNegotiation("application/json")).Get("/{uuid}/files", c.GetFile)
				r.Delete("/{uuid}/files/*", c.DeleteFile)
				// the following routes always return an error but are included for API consistency
				r.With(middleware.RequireContentType("application/octet-stream")).Put("/{uuid}/files", c.PutFile)
				r.Delete("/{uuid}/files", c.DeleteFile)
			})

			r.Group(func(r chi.Router) {
				r.Use(UploadState(model.UploadStateProcessing, model.UploadStateDone))
				r.With(middleware.Paginate(100, 1000), render.ContentTypeNegotiation("application/json")).Get("/{uuid}/errors", c.GetErrors)
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
}
