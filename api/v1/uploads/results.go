package uploads

import (
	"net/http"

	"github.com/lmittmann/tint"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// GetErrors implements the GET /v1/uploads/{uuid}/errors endpoint.
func (h *Handler) GetErrors(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	pagination := middleware.MustGetPagination(r.Context())
	errors, total, err := h.uploadRepo.GetErrors(r.Context(), upload.UUID, pagination.Limit, pagination.Offset)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Could not load upload errors.", "uuid", upload.UUID, "limit", pagination.Limit, "offset", pagination.Offset, tint.Err(err))
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}

	resp := schema.List[*schema.UploadProcessingError]{
		Items:  make([]*schema.UploadProcessingError, len(errors)),
		Offset: pagination.Offset,
		Limit:  pagination.RequestLimit,
		Total:  total,
	}
	for i, errVal := range errors {
		s := schema.FromUploadProcessingError(errVal)
		resp.Items[i] = &s
	}
	_ = render.Render(w, r, &resp)
}
