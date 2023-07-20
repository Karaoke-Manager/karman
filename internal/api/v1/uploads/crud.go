package uploads

import (
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/api/middleware"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	upload, err := c.Service.CreateUpload(r.Context())
	if err != nil {
		// FIXME: Are there special other errors that we should handle?
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}

	resp := schema.NewUploadFromModel(upload)
	_ = render.Render(w, r, resp)
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	resp := schema.NewUploadFromModel(upload)
	_ = render.Render(w, r, resp)
}

func (c *Controller) Find(w http.ResponseWriter, r *http.Request) {
	pagination := middleware.MustGetPagination(r.Context())
	// TODO: Do limit-offset pagination
	uploads, total, err := c.Service.FindUploads(r.Context(), pagination.Limit, pagination.Offset)
	if err != nil {
		// FIXME: Differentiate other errors?
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}

	uploadSchemas := make([]*schema.Upload, len(uploads))
	for i, upload := range uploads {
		uploadSchemas[i] = schema.NewUploadFromModel(upload)
	}
	resp := &schema.List[*schema.Upload]{
		Items:  uploadSchemas,
		Offset: pagination.Offset,
		Limit:  pagination.Limit,
		Total:  total,
	}
	_ = render.Render(w, r, resp)
}

func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	if err := c.Service.DeleteUploadByUUID(r.Context(), uuid); err != nil {
		_ = render.Render(w, r, apierror.DBError(err))
		return
	}
	_ = render.NoContent(w, r)
}
