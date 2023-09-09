package songs

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/core/media"
	"github.com/Karaoke-Manager/karman/core/song"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// Handler implements the /v1/songs endpoints.
type Handler struct {
	logger *slog.Logger
	r      chi.Router

	songRepo   song.Repository
	songSvc    song.Service
	mediaStore media.Store
	mediaSvc   media.Service
}

// NewHandler creates a new Handler instance using the specified services.
func NewHandler(logger *slog.Logger, songRepo song.Repository, songSvc song.Service, mediaStore media.Store, mediaSvc media.Service) *Handler {
	r := chi.NewRouter()
	h := &Handler{logger.With("log", "songs.handler"), r, songRepo, songSvc, mediaStore, mediaSvc}

	r.With(middleware.RequireContentType("text/plain", "text/x-ultrastar"), render.ContentTypeNegotiation("application/json")).Post("/", h.Create)
	r.With(middleware.Paginate(25, 100), render.ContentTypeNegotiation("application/json")).Get("/", h.Find)

	r.Group(func(r chi.Router) {
		r.Use(middleware.UUID("uuid"))
		r.Delete("/{uuid}", h.Delete)

		r.Group(func(r chi.Router) {
			r.Use(h.FetchSong)
			r.With(render.ContentTypeNegotiation("application/json")).Get("/{uuid}", h.Get)
			r.With(render.ContentTypeNegotiation("text/x-ultrastar", "text/plain")).Get("/{uuid}/txt", h.GetTxt)
			// r.Get("{uuid}/archive", h.GetArchive)
			r.With(render.ContentTypeNegotiation("image/*")).Get("/{uuid}/cover", h.GetCover)
			r.With(render.ContentTypeNegotiation("image/*")).Get("/{uuid}/background", h.GetBackground)
			r.With(render.ContentTypeNegotiation("audio/*")).Get("/{uuid}/audio", h.GetAudio)
			r.With(render.ContentTypeNegotiation("video/*")).Get("/{uuid}/video", h.GetVideo)

			// Deleting media is allowed in uploads
			r.Delete("/{uuid}/cover", h.DeleteCover)
			r.Delete("/{uuid}/background", h.DeleteBackground)
			r.Delete("/{uuid}/audio", h.DeleteAudio)
			r.Delete("/{uuid}/video", h.DeleteVideo)
		})

		r.Group(func(r chi.Router) {
			r.Use(h.FetchSong, h.CheckModify)
			r.With(middleware.ContentTypeJSON).Patch("/{uuid}", h.Update)
			r.With(middleware.RequireContentType("text/plain", "text/x-ultrastar"), render.ContentTypeNegotiation("application/json")).Put("/{uuid}/txt", h.ReplaceTxt)
			r.With(middleware.RequireContentType("image/*")).Put("/{uuid}/cover", h.ReplaceCover)
			r.With(middleware.RequireContentType("image/*")).Put("/{uuid}/background", h.ReplaceBackground)
			r.With(middleware.RequireContentType("audio/*")).Put("/{uuid}/audio", h.ReplaceAudio)
			r.With(middleware.RequireContentType("video/*")).Put("/{uuid}/video", h.ReplaceVideo)
		})
	})
	return h
}

// ServeHTTP processes HTTP requests for h.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}
