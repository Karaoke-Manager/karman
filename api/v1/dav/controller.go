package dav

import (
	"github.com/go-chi/chi/v5"
	"golang.org/x/net/webdav"

	"github.com/Karaoke-Manager/karman/api/v1/dav/internal"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
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

// Controller implements the /v1/dav endpoints.
type Controller struct {
	songSvc  song.Service
	mediaSvc media.Service
}

// NewController creates a new controller instance using the specified services.
func NewController(songService song.Service, mediaService media.Service) *Controller {
	return &Controller{songService, mediaService}
}

// Router sets up the routing for the endpoint.
func (c *Controller) Router(r chi.Router) {
	h := &webdav.Handler{
		// TODO: Make this configurable/dynamic
		Prefix:     "/api/v1/dav/",
		FileSystem: internal.NewFlatFS(c.songSvc, c.mediaSvc),
		LockSystem: webdav.NewMemLS(),
		Logger:     nil,
	}
	r.Handle("/*", h)
}
