package uploads

import (
	"net/http"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// GetErrors implements the GET /v1/uploads/{uuid}/errors endpoint.
func (c *Controller) GetErrors(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	pagination := middleware.MustGetPagination(r.Context())
	errors, total, err := c.svc.GetErrors(r.Context(), upload, pagination.Limit, pagination.Offset)
	if err != nil {
		_ = render.Render(w, r, apierror.ServiceError(err))
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
