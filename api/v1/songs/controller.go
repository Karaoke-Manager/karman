package songs

import (
	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/pkg/render"
	_ "github.com/Karaoke-Manager/karman/pkg/render/json"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
)

// Controller implements the /v1/songs endpoints.
type Controller struct {
	songRepo song.Repository
	mediaSvc media.Service
}

// NewController creates a new Controller instance using the specified services.
func NewController(songRepo song.Repository, mediaSvc media.Service) *Controller {
	return &Controller{songRepo, mediaSvc}
}

// Router sets up the routing for the endpoint.
func (c *Controller) Router(r chi.Router) {
	r.With(middleware.RequireContentType("text/plain", "text/x-ultrastar"), render.ContentTypeNegotiation("application/json")).Post("/", c.Create)
	r.With(middleware.Paginate(25, 100), render.ContentTypeNegotiation("application/json")).Get("/", c.Find)

	r.Group(func(r chi.Router) {
		r.Use(middleware.UUID("uuid"))
		r.Delete("/{uuid}", c.Delete)

		r.Group(func(r chi.Router) {
			r.Use(c.FetchSong)
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
			r.Use(c.FetchSong, c.CheckModify)
			r.With(middleware.ContentTypeJSON).Patch("/{uuid}", c.Update)
			r.With(middleware.RequireContentType("text/plain", "text/x-ultrastar"), render.ContentTypeNegotiation("application/json")).Put("/{uuid}/txt", c.ReplaceTxt)
			r.With(middleware.RequireContentType("image/*")).Put("/{uuid}/cover", c.ReplaceCover)
			r.With(middleware.RequireContentType("image/*")).Put("/{uuid}/background", c.ReplaceBackground)
			r.With(middleware.RequireContentType("audio/*")).Put("/{uuid}/audio", c.ReplaceAudio)
			r.With(middleware.RequireContentType("video/*")).Put("/{uuid}/video", c.ReplaceVideo)
		})
	})
}
