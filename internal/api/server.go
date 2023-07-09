package api

import (
	"github.com/Karaoke-Manager/karman/internal/api/apiv1"
	"github.com/Karaoke-Manager/karman/pkg/rwfs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type Server struct {
	v1Server *apiv1.Server
}

func NewServer(db *gorm.DB, songFS rwfs.FS, uploadFS rwfs.FS) *Server {
	s := &Server{
		v1Server: apiv1.NewServer(db, songFS, uploadFS),
	}
	return s
}

func (s *Server) Router(r chi.Router) {
	// Restrict requests to JSON for now
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.CleanPath)
	// Some CORS stuff
	// r.Use(middleware.Compress())
	// r.Use(middleware.RealIP)
	// r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Route("/v1", s.v1Server.Router)

	r.NotFound(s.NotFound)
	r.MethodNotAllowed(s.MethodNotAllowed)
}
