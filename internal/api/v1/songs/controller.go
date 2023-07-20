package songs

import (
	"github.com/Karaoke-Manager/karman/internal/api/middleware"
	"github.com/Karaoke-Manager/karman/internal/service/song"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	svc song.Service
}

func NewController(svc song.Service) *Controller {
	s := &Controller{svc}
	return s
}

func (c *Controller) Router(r chi.Router) {
	r.With(middleware.RequireContentType("text/plain")).Post("/", c.Create)
	r.With(middleware.Paginate(25, 100)).Get("/", c.Find)

	r.Group(func(r chi.Router) {
		r.Use(c.fetchUpload)
		r.Get("/{uuid}", c.Get)
	})
	// GET /{uuid}
	// POST /{uuid}
	// PATCH /{uuid}
	// DELETE /{uuid}

	// GET /{uuid}/artwork
	// POST /{uuid}/artwork (JSON, image types)
	// GET /{uuid}/audio
	// POST /{uuid}/audio (JSON, audio types)
	// GET /{uuid}/video
	// POST /{uuid}/video (JSON, video types)
	// GET /{uuid}/background
	// POST /{uuid}/background (JSON, image types)
	// GET /{uuid}/txt?????
	// POST /{uuid}/txt (txt, file references are ignored)
}
