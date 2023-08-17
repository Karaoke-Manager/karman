package uploads

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	schema2 "github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	upload, err := c.Service.CreateUpload(r.Context())
	if err != nil {
		// FIXME: Are there special other errors that we should handle?
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}

	resp := schema2.NewUploadFromModel(upload)
	_ = render.Render(w, r, resp)
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	resp := schema2.NewUploadFromModel(upload)
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

	uploadSchemas := make([]*schema2.Upload, len(uploads))
	for i, upload := range uploads {
		uploadSchemas[i] = schema2.NewUploadFromModel(upload)
	}
	resp := &schema2.List[*schema2.Upload]{
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