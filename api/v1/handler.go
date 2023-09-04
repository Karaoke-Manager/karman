package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/karman/api/v1/dav"
	"github.com/Karaoke-Manager/karman/api/v1/songs"
	"github.com/Karaoke-Manager/karman/api/v1/uploads"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
	"github.com/Karaoke-Manager/karman/service/upload"
)

// Handler implements the /v1 API namespace.
type Handler struct {
	r chi.Router
}

// NewHandler creates a new handler using the specified services.
// This function will create the required sub-handlers automatically.
func NewHandler(songRepo song.Repository, songSvc song.Service, mediaSvc media.Service, mediaStore media.Store, uploadRepo upload.Repository, uploadStore upload.Store) *Handler {
	uploadsHandler := uploads.NewHandler(uploadRepo, uploadStore)
	songsHandler := songs.NewHandler(songRepo, songSvc, mediaStore, mediaSvc)
	davHandler := dav.NewHandler(songRepo, songSvc, mediaStore)

	r := chi.NewRouter()
	h := &Handler{r}
	r.Mount("/songs", songsHandler)
	r.Mount("/uploads", uploadsHandler)
	r.Mount("/dav", davHandler)
	return h
}

// ServeHTTP processes HTTP requests for h.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}