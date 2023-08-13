package songs

import (
	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/karman/internal/api/middleware"
	"github.com/Karaoke-Manager/karman/internal/service/media"
	"github.com/Karaoke-Manager/karman/internal/service/song"
	"github.com/Karaoke-Manager/karman/pkg/render"
	_ "github.com/Karaoke-Manager/karman/pkg/render/json"
)

// Controller implements the /v1/songs endpoint.
type Controller struct {
	songSvc  song.Service
	mediaSvc media.Service
}

// NewController creates a new Controller instance using the specified services.
func NewController(songSvc song.Service, mediaSvc media.Service) *Controller {
	return &Controller{songSvc, mediaSvc}
}

// Router sets up the routing for the endpoint.
func (c *Controller) Router(r chi.Router) {
	r.With(middleware.RequireContentType("text/plain", "text/x-ultrastar")).Post("/", c.Create)
	r.With(middleware.Paginate(25, 100)).Get("/", c.Find)
	r.With(middleware.UUID("uuid")).Delete("/{uuid}", c.Delete)

	r.Group(func(r chi.Router) {
		r.Use(middleware.UUID("uuid"))

		r.Group(func(r chi.Router) {
			r.Use(c.FetchUpload)
			r.With(render.ContentTypeNegotiation("application/json")).Get("/{uuid}", c.Get)
			r.With(render.ContentTypeNegotiation("text/x-ultrastar", "text/plain")).Get("/{uuid}/txt", c.GetTxt)
			// r.Get("{uuid}/archive", c.GetArchive)
			r.With(render.ContentTypeNegotiation("image/*")).Get("/{uuid}/cover", c.GetCover)
			r.With(render.ContentTypeNegotiation("image/*")).Get("/{uuid}/background", c.GetBackground)
			r.With(render.ContentTypeNegotiation("audio/*")).Get("/{uuid}/audio", c.GetAudio)
			r.With(render.ContentTypeNegotiation("video/*")).Get("/{uuid}/video", c.GetVideo)

			// Deleting media is allowed in uploads
			r.Delete("/{uuid}/cover", c.DeleteCover)
			r.Delete("/{uuid}/background", c.DeleteBackground)
			r.Delete("/{uuid}/audio", c.DeleteAudio)
			r.Delete("/{uuid}/video", c.DeleteVideo)
		})

		r.Group(func(r chi.Router) {
			r.Use(c.FetchUpload, c.CheckModify)
			r.With(middleware.ContentTypeJSON).Patch("/{uuid}", c.Update)
			r.With(middleware.RequireContentType("text/plain", "text/x-ultrastar"), render.ContentTypeNegotiation("application/json")).Put("/{uuid}/txt", c.ReplaceTxt)
			r.With(middleware.RequireContentType("image/*")).Put("/{uuid}/cover", c.ReplaceCover)
			r.With(middleware.RequireContentType("image/*")).Put("/{uuid}/background", c.ReplaceBackground)
			r.With(middleware.RequireContentType("audio/*")).Put("/{uuid}/audio", c.ReplaceAudio)
			r.With(middleware.RequireContentType("video/*")).Put("/{uuid}/video", c.ReplaceVideo)
		})
	})
}
