package upload

import (
	"github.com/Karaoke-Manager/karman/internal/apierror"
	"github.com/Karaoke-Manager/karman/internal/models"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"io/fs"
	"net/http"
)

type FileSchema struct {
	render.NopRenderer
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	IsDir bool   `json:"directory"`
}

type GetFileResponseSchema struct {
	render.NopRenderer
	UploadUUID uuid.UUID    `json:"upload"`
	File       FileSchema   `json:"file"`
	Children   []FileSchema `json:"children,omitempty"`
}

func (s *Server) GetFile(w http.ResponseWriter, r *http.Request) {
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
	stat, err := fs.Stat(uploadFS, path)
	if err != nil {
		_ = apierror.NotFound(w, r)
		return
	}
	resp := GetFileResponseSchema{
		UploadUUID: upload.UUID,
		File: FileSchema{
			Name:  stat.Name(),
			Size:  stat.Size(),
			IsDir: stat.IsDir(),
		},
	}
	if stat.IsDir() {
		entries, err := fs.ReadDir(uploadFS, path+"/"+stat.Name())
		if err != nil {
			_ = apierror.InternalServerError(w, r)
			return
		}
		children := make([]FileSchema, len(entries))
		for i, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				_ = apierror.InternalServerError(w, r)
				return
			}
			children[i] = FileSchema{
				Name:  entry.Name(),
				Size:  info.Size(),
				IsDir: entry.IsDir(),
			}
		}
		resp.Children = children
	}
	_ = render.Render(w, r, resp)
}
