package uploads

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/test"
)

func TestController_GetErrors(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, uploadPath(data.UploadWithErrors, "/errors"), nil)
		resp := test.DoRequest(h, r)

		test.AssertPagination(t, resp, 0, 100, 2, 2)
		var errors []schema.UploadProcessingError
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&errors), "decode errors") {
			assert.Len(t, errors, 2, "length of result")
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/errors"))
	t.Run("400 Bad Request (Pagination)", test.InvalidPagination(h, http.MethodGet, "/"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodGet, uploadPath(data.AbsentUpload, "/errors"), http.StatusNotFound))
	t.Run("409 Conflict", testInvalidState(h, http.MethodGet, uploadPath(data.OpenUpload, "/errors"), data.OpenUpload.UUID))
}
