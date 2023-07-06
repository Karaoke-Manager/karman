package apiv1

import (
	"github.com/Karaoke-Manager/karman/internal/resources/upload"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"io/fs"
)

type Server struct {
	uploadServer *upload.Server
}

func NewServer(db *gorm.DB, uploadFS fs.FS) *Server {
	s := &Server{
		uploadServer: upload.NewServer(db, uploadFS),
	}
	return s
}

func (s *Server) Router(r chi.Router) {
	r.Route("/uploads", s.uploadServer.Router)
}
