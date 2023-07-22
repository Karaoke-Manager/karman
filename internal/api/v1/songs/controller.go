package songs

import (
	"github.com/Karaoke-Manager/karman/internal/api/middleware"
	"github.com/Karaoke-Manager/karman/internal/service/song"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	svc song.Service
}

func NewController(svc song.Service) Controller {
	return Controller{svc}
}

func (c Controller) Router(r chi.Router) {
	r.With(middleware.RequireContentType("text/plain")).Post("/", c.Create)
	r.With(middleware.Paginate(25, 100)).Get("/", c.Find)
	r.With(middleware.UUID("uuid")).Delete("/{uuid}", c.Delete)

	r.Group(func(r chi.Router) {
		r.Use(middleware.UUID("uuid"))

		r.Group(func(r chi.Router) {
			r.Use(c.FetchUpload(false))
			r.With(c.CheckModify).With(middleware.ContentTypeJSON).Patch("/{uuid}", c.Update)
		})

		r.Group(func(r chi.Router) {
			r.Use(c.FetchUpload(true))
			r.Get("/{uuid}", c.Get)
			r.Get("/{uuid}/txt", c.GetTxt)
			r.With(c.CheckModify).With(middleware.RequireContentType("text/plain")).Put("/{uuid}/txt", c.ReplaceTxt)
		})
	})

	// GET /{uuid}/artwork
	// POST /{uuid}/artwork (JSON, image types)
	// GET /{uuid}/audio
	// POST /{uuid}/audio (JSON, audio types)
	// GET /{uuid}/video
	// POST /{uuid}/video (JSON, video types)
	// GET /{uuid}/background
	// POST /{uuid}/background (JSON, image types)
	// POST /{uuid}/txt (txt, file references are ignored)
}
