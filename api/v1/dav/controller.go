package dav

import (
	"github.com/go-chi/chi/v5"
	"golang.org/x/net/webdav"

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

type Controller struct {
	songSvc  song.Service
	mediaSvc media.Service
}

func NewController(songService song.Service, mediaService media.Service) *Controller {
	return &Controller{songService, mediaService}
}

func (c *Controller) Router(r chi.Router) {
	h := &webdav.Handler{
		// TODO: Make this configurable/dynamic
		Prefix:     "/api/v1/dav/",
		FileSystem: NewFlatFS(c.songSvc, c.mediaSvc),
		LockSystem: webdav.NewMemLS(),
		Logger:     nil,
	}
	r.Handle("/*", h)
}
