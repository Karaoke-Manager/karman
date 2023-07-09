package apiv1

import (
	"github.com/Karaoke-Manager/karman/internal/resources/song"
	"github.com/Karaoke-Manager/karman/internal/resources/upload"
	"github.com/Karaoke-Manager/karman/pkg/rwfs"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Server struct {
	uploadServer *upload.Server
	songServer   *song.Server
}

func NewServer(db *gorm.DB, songFS rwfs.FS, uploadFS rwfs.FS) *Server {
	s := &Server{
		uploadServer: upload.NewServer(db, uploadFS),
		songServer:   song.NewServer(db, songFS),
	}
	return s
}

func (s *Server) Router(r chi.Router) {
	r.Route("/songs", s.songServer.Router)
	r.Route("/uploads", s.uploadServer.Router)
}
