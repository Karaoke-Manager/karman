package uploads

import (
	"errors"
	"io"
	"io/fs"
	"net/http"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/Karaoke-Manager/karman/service/upload"
)

// PutFile implements the PUT /v1/uploads/{uuid}/files/* endpoint.
func (h *Handler) PutFile(w http.ResponseWriter, r *http.Request) {
	u := MustGetUpload(r.Context())
	path := MustGetFilePath(r.Context())
	if path == "." {
		_ = render.Render(w, r, apierror.InvalidUploadPath("."))
		return
	}
	f, err := h.uploadStore.Create(r.Context(), u.UUID, path)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_, err = io.Copy(f, r.Body)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	err = f.Close()
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// GetFile implements the GET /v1/uploads/{uuid}/files/* endpoint.
func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {
	u := MustGetUpload(r.Context())
	path := MustGetFilePath(r.Context())
	marker := r.URL.Query().Get("marker")

	stat, err := h.uploadStore.Stat(r.Context(), u.UUID, path)
	if errors.Is(err, fs.ErrNotExist) {
		_ = render.Render(w, r, apierror.UploadFileNotFound(u, path))
		return
	}
	var children []fs.FileInfo
	if stat.IsDir() {
		f, err := h.uploadStore.Open(r.Context(), u.UUID, path)
		if err != nil {
			_ = render.Render(w, r, apierror.ErrInternalServerError)
		}
		dir := f.(upload.Dir)
		if err = dir.SkipTo(marker); err != nil {
			_ = render.Render(w, r, apierror.ErrInternalServerError)
			return
		}
		children, err = dir.Readdir(500)
		if errors.Is(err, io.EOF) {
			marker = ""
		} else if err != nil {
			_ = render.Render(w, r, apierror.ErrInternalServerError)
			return
		} else {
			marker = dir.Marker()
		}
	}
	s := schema.FromUploadFileStat(stat, children, marker)
	if path == "." {
		// Do not include root dir name
		s.Name = ""
	}
	_ = render.Render(w, r, s)
}

// DeleteFile implements the DELETE /v1/uploads/{uuid}/files/* endpoint.
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	u := MustGetUpload(r.Context())
	path := MustGetFilePath(r.Context())
	if path == "." {
		_ = render.Render(w, r, apierror.InvalidUploadPath("."))
		return
	}
	if err := h.uploadStore.Delete(r.Context(), u.UUID, path); err != nil {
		_ = render.Render(w, r, apierror.ServiceError(err))
		return
	}
	_ = render.NoContent(w, r)
}
