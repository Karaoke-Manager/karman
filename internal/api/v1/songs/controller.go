package songs

import (
	"github.com/Karaoke-Manager/karman/pkg/rwfs"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Controller struct {
	db *gorm.DB
	fs rwfs.FS
}

func NewController(db *gorm.DB, filesystem rwfs.FS) *Controller {
	s := &Controller{db, filesystem}
	return s
}

func (c *Controller) Router(r chi.Router) {
	// GET / (list)
	// GET /{uuid}
	// POST /{uuid}
	// PATCH /{uuid}
	// DELETE /{uuid}

	// GET /{uuid}/artwork
	// POST /{uuid}/artwork (JSON, image types)
	// GET /{uuid}/audio
	// POST /{uuid}/audio (JSON, audio types)
	// GET /{uuid}/video
	// POST /{uuid}/video (JSON, video types)
	// GET /{uuid}/background
	// POST /{uuid}/background (JSON, image types)
	// GET /{uuid}/txt?????
	// POST /{uuid}/txt (txt, file references are ignored)
}
