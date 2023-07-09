package upload

import (
	"github.com/Karaoke-Manager/karman/internal/apierror"
	"github.com/Karaoke-Manager/karman/internal/models"
	"github.com/go-chi/chi/v5"
	"io/fs"
	"net/http"
)

func (s *Server) PutFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "uuid")
	path := chi.URLParam(r, "*")
	var upload models.Upload
	if err := s.db.First(&upload, "uuid = ?", id).Error; err != nil {
		_ = apierror.DBError(w, r, err)
		return
	}
	// TODO: This should be a chroot style sub, that forbids breakout via symlinks
	uploadFS, err := fs.Sub(s.fs, upload.UUID.String())
	if err != nil {
		_ = apierror.InternalServerError(w, r)
		return
	}
}
