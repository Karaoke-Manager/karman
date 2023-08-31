package uploads

import (
	"net/http"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// Create implements the POST /v1/uploads endpoint.
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	upload, err := c.svc.CreateUpload(r.Context())
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}

	resp := schema.FromUpload(upload)
	render.SetStatus(r, http.StatusCreated)
	_ = render.Render(w, r, resp)
}

// Find implements the GET /v1/uploads endpoint.
func (c *Controller) Find(w http.ResponseWriter, r *http.Request) {
	pagination := middleware.MustGetPagination(r.Context())
	uploads, total, err := c.svc.FindUploads(r.Context(), pagination.Limit, pagination.Offset)
	if err != nil {
		_ = render.Render(w, r, apierror.ServiceError(err))
		return
	}

	resp := schema.List[*schema.Upload]{
		Items:  make([]*schema.Upload, len(uploads)),
		Offset: pagination.Offset,
		Limit:  pagination.RequestLimit,
		Total:  total,
	}
	for i, upload := range uploads {
		s := schema.FromUpload(upload)
		resp.Items[i] = &s
	}
	_ = render.Render(w, r, &resp)
}

// Get implements the GET /v1/uploads/{uuid} endpoint.
func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	resp := schema.FromUpload(upload)
	_ = render.Render(w, r, &resp)
}

// Delete implements the DELETE /v1/uploads/{uuid} endpoint.
func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	id := middleware.MustGetUUID(r.Context())
	if err := c.svc.DeleteUpload(r.Context(), id); err != nil {
		_ = render.Render(w, r, apierror.ServiceError(err))
		return
	}
	_ = render.NoContent(w, r)
}
