package uploads

import (
	"net/http"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// Create implements the POST /v1/uploads endpoint.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	upload := model.Upload{}
	if err := h.uploadRepo.CreateUpload(r.Context(), &upload); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}

	resp := schema.FromUpload(upload)
	render.SetStatus(r, http.StatusCreated)
	_ = render.Render(w, r, resp)
}

// Find implements the GET /v1/uploads endpoint.
func (h *Handler) Find(w http.ResponseWriter, r *http.Request) {
	pagination := middleware.MustGetPagination(r.Context())
	uploads, total, err := h.uploadRepo.FindUploads(r.Context(), pagination.Limit, pagination.Offset)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
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
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	resp := schema.FromUpload(upload)
	_ = render.Render(w, r, &resp)
}

// Delete implements the DELETE /v1/uploads/{uuid} endpoint.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := middleware.MustGetUUID(r.Context())
	if _, err := h.uploadRepo.DeleteUpload(r.Context(), id); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}
