package song

import (
	"github.com/Karaoke-Manager/karman/pkg/rwfs"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Server struct {
	db *gorm.DB
	fs rwfs.FS
}

func NewServer(db *gorm.DB, filesystem rwfs.FS) *Server {
	s := &Server{db, filesystem}
	return s
}

func (s *Server) Router(r chi.Router) {
	// GET / (list)
	// GET /{uuid}
	// POST /{uuid}
	// PATCH /{uuid}
	// DELETE /{uuid}

	// GET /{uuid}/artwork
	// GET /{uuid}/audio
	// GET /{uuid}/video
	// GET /{uuid}/background
	// GET /{uuid}/txt?????
}
