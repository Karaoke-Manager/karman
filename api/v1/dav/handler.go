package dav

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/webdav"

	"github.com/Karaoke-Manager/karman/api/v1/dav/internal"
	"github.com/Karaoke-Manager/karman/core/media"
	"github.com/Karaoke-Manager/karman/core/song"
)

func init() {
	// see https://de.wikipedia.org/wiki/WebDAV
	chi.RegisterMethod("PROPFIND")
	chi.RegisterMethod("PROPPATCH")
	chi.RegisterMethod("MKCOL")
	chi.RegisterMethod("COPY")
	chi.RegisterMethod("MOVE")
	chi.RegisterMethod("DELETE")
	chi.RegisterMethod("LOCK")
	chi.RegisterMethod("UNLOCK")
}

// Handler implements the /v1/dav endpoints.
type Handler struct {
	wh *webdav.Handler
}

// NewHandler creates a new Handler instance using the specified services.
func NewHandler(
	logger *slog.Logger,
	songRepo song.Repository,
	songSvc song.Service,
	mediaStore media.Store,
) *Handler {
	wh := &webdav.Handler{
		// TODO: Make this configurable/dynamic
		Prefix:     "/v1/dav/",
		FileSystem: internal.NewFlatFS(logger, songRepo, songSvc, mediaStore),
		LockSystem: webdav.NewMemLS(),
		Logger:     nil,
	}
	return &Handler{wh}
}

// ServeHTTP processes HTTP requests for h.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.wh.ServeHTTP(w, r)
}
