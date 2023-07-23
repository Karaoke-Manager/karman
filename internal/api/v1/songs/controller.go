package songs

import (
	"github.com/Karaoke-Manager/karman/internal/api/middleware"
	"github.com/Karaoke-Manager/karman/internal/service/media"
	"github.com/Karaoke-Manager/karman/internal/service/song"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	songSvc  song.Service
	mediaSvc media.Service
}

func NewController(songSvc song.Service, mediaSvc media.Service) Controller {
	return Controller{songSvc, mediaSvc}
}

func (c Controller) Router(r chi.Router) {
	r.With(middleware.RequireContentType("text/plain")).Post("/", c.Create)
	r.With(middleware.Paginate(25, 100)).Get("/", c.Find)
	r.With(middleware.UUID("uuid")).Delete("/{uuid}", c.Delete)

	r.Group(func(r chi.Router) {
		r.Use(middleware.UUID("uuid"))

		r.Group(func(r chi.Router) {
			r.Use(c.FetchUpload(true))
			r.Get("/{uuid}", c.Get)
			r.Get("/{uuid}/txt", c.GetTxt)
			// r.Get("{uuid}/archive", c.GetArchive)
			r.Get("/{uuid}/cover", c.GetCover)
		})

		r.Group(func(r chi.Router) {
			r.Use(c.FetchUpload(false), c.CheckModify)
			r.With(middleware.ContentTypeJSON).Patch("/{uuid}", c.Update)
			r.With(middleware.RequireContentType("text/plain")).Put("/{uuid}/txt", c.ReplaceTxt)
			r.With(middleware.RequireContentType("image/*")).Put("/{uuid}/cover", c.ReplaceCover)
		})

		r.Group(func(r chi.Router) {
			r.Use(c.FetchUpload(false))
			// Deleting media is allowed in uploads
			r.Delete("/{uuid}/cover", c.DeleteCover)
		})
	})

	// GET /{uuid}/audio
	// PUT /{uuid}/audio
	// DELETE /{uuid}/audio
	// GET /{uuid}/video
	// PUT /{uuid}/video
	// DELETE /{uuid}/video
	// GET /{uuid}/background
	// PUT /{uuid}/background
	// DELETE /{uuid}/background
}
