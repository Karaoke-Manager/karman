//go:build database

package uploads

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func TestHandler_GetErrors(t *testing.T) {
	t.Parallel()

	h, db := setupHandler(t, "/v1/uploads/")
	openUpload := testdata.OpenUpload(t, db)
	uploadWithErrors := testdata.DoneUploadWithErrors(t, db)

	t.Run("200 OK", func(t *testing.T) {
		url := fmt.Sprintf("/v1/uploads/%s/errors", uploadWithErrors.UUID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose

		test.AssertPagination(t, resp, 0, 100, uploadWithErrors.Errors, int64(uploadWithErrors.Errors))
		var errors []schema.UploadProcessingError
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if err := json.NewDecoder(resp.Body).Decode(&errors); err != nil {
			t.Errorf("GET %s responded with invalid upload error schema: %s", url, err)
			return
		}
		if len(errors) != uploadWithErrors.Errors {
			t.Errorf("GET %s responded with %d errors, expected %d", url, len(errors), uploadWithErrors.Errors)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/uploads/%s/errors", testdata.InvalidUUID)))
	t.Run("400 Bad Request (Pagination)", test.InvalidPagination(h, http.MethodGet, fmt.Sprintf("/v1/uploads/%s/errors", uploadWithErrors.UUID)))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodGet, fmt.Sprintf("/v1/uploads/%s/errors", uuid.New()), http.StatusNotFound))
	t.Run("409 Conflict", testInvalidState(h, http.MethodGet, "/v1/uploads/%s/errors", openUpload.UUID))
}
