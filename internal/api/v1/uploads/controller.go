package uploads

import (
	"github.com/Karaoke-Manager/karman/internal/service/upload"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Controller struct {
	upload.Service
}

func NewController(svc upload.Service) *Controller {
	s := &Controller{svc}
	return s
}

func (c *Controller) Router(r chi.Router) {
	r.Get("/", c.List)
	r.Post("/", c.Create)
	r.Delete("/{uuid}", c.Delete)

	r.Group(func(r chi.Router) {
		r.Use(c.fetchUpload)

		r.Get("/{uuid}", c.Get)

		r.Group(func(r chi.Router) {
			r.Use(c.validateFilePath)

			r.Get("/{uuid}/files/*", c.GetFile)
			// FIXME: Stacking allow content type middleware like that does not work.
			// FIXME: The response by this middleware does not fit our error types.
			r.With(middleware.AllowContentType("application/octet-stream")).Put("/{uuid}/files/*", c.PutFile)
			r.Delete("/{uuid}/files/*", c.DeleteFile)
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
	})
}
