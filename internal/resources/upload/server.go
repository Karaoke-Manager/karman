package upload

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"io/fs"
)

type Server struct {
	db *gorm.DB
	fs fs.FS
}

func NewServer(db *gorm.DB, filesystem fs.FS) *Server {
	s := &Server{db, filesystem}
	return s
}

func (s *Server) Router(r chi.Router) {
	r.Get("/", s.List)
	r.Post("/", s.Create)
	r.Get("/{uuid}", s.Get)
	r.Delete("/{uuid}", s.Delete)

	// GET /{uuid}/files/{*path/to/file/or/folder}
	// POST {uuid}/files/{*path/to/file.mp3}

	// GET /{uuid}/songs

	// OPTION 1:
	// GET /{uuid}/songs/{id2}
	// DELETE /{uuid}/songs{id2}
	// POST /{uuid}/import

	// OPTION 2:
	// GET /{uuid}/songs
	// GET /songs/...
	// DELETE /songs/...
	// POST /{uuid}/import (with the songs to import)
}
