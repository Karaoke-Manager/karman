package uploads

import (
	"errors"
	"io"
	"io/fs"
	"net/http"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

func (c *Controller) PutFile(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	path := MustGetFilePath(r.Context())
	f, err := c.svc.CreateFile(r.Context(), upload, path)
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

func (c *Controller) GetFile(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	path := MustGetFilePath(r.Context())
	marker := r.URL.Query().Get("marker")

	stat, err := c.svc.StatFile(r.Context(), upload, path)
	if errors.Is(err, fs.ErrNotExist) {
		_ = render.Render(w, r, apierror.UploadFileNotFound(upload, path))
		return
	}
	var children []fs.FileInfo
	if stat.IsDir() {
		dir, err := c.svc.OpenDir(r.Context(), upload, path)
		if err != nil {
			_ = render.Render(w, r, apierror.ErrInternalServerError)
		}
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
	_ = render.Render(w, r, s)
}

/*
func (c *Controller) handleFileError(w http.ResponseWriter, r *http.Request, upload *model.Upload, path string, err error) {
	var details *apierror.ProblemDetails
	switch {
	case errors.Is(err, uploadSvc.ErrUploadClosed):
		details = apierror.UploadClosed(upload)
	case errors.Is(err, fs.ErrNotExist):
		details = apierror.UploadFileNotFound(upload, path)
	default:
		details = apierror.ErrInternalServerError
	}
	_ = render.Render(w, r, details)
}

func (c *Controller) DeleteFile(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	path := MustGetFilePath(r.Context())

	if err := c.Service.DeleteFile(r.Context(), upload, path); err != nil {
		c.handleFileError(w, r, upload, path, err)
		return
	}
	_ = render.NoContent(w, r)
}
*/
