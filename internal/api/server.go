package api

import (
	"github.com/Karaoke-Manager/karman/internal/api/apiv1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
	"io/fs"
)

type Server struct {
	v1Server *apiv1.Server
}

func NewServer(db *gorm.DB, uploadFS fs.FS) *Server {
	s := &Server{
		v1Server: apiv1.NewServer(db, uploadFS),
	}
	return s
}

func (s *Server) Router(r chi.Router) {
	r.Use(middleware.AllowContentType("application/json"))
	r.Route("/v1", s.v1Server.Router)

	r.NotFound(s.NotFound)
	r.MethodNotAllowed(s.MethodNotAllowed)
}
