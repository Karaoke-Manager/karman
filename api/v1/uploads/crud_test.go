package uploads

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/test"
)

func TestController_Create(t *testing.T) {
	h, _, _ := setup(t, false)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		resp := test.DoRequest(h, r)

		var upload schema.Upload
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&upload), "response decode")
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "status Code")
		assert.NotEmpty(t, upload.UUID)
		assert.Equal(t, model.UploadStateOpen, upload.Status)
	})
}

func TestController_Find(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		resp := test.DoRequest(h, r)

		test.AssertPagination(t, resp, 0, 25, int(data.TotalUploads), data.TotalUploads)
		var uploads []schema.Upload
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&uploads), "decode uploads") {
			assert.Len(t, uploads, int(data.TotalUploads), "length of result")
		}
	})
	t.Run("400 Bad Request (Pagination)", test.InvalidPagination(h, http.MethodGet, "/"))
}

func TestController_Get(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK (Open)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/"+data.OpenUpload.UUID.String(), nil)
		resp := test.DoRequest(h, r)

		var upload schema.Upload
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&upload), "decode upload") {
			assert.Equal(t, data.OpenUpload.UUID, upload.UUID, "upload UUID")
			assert.Equal(t, data.OpenUpload.State, upload.Status, "upload status")
		}
	})
	t.Run("200 OK (Pending)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/"+data.PendingUpload.UUID.String(), nil)
		resp := test.DoRequest(h, r)

		var upload schema.Upload
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&upload), "decode upload") {
			assert.Equal(t, data.PendingUpload.UUID, upload.UUID, "upload UUID")
			assert.Equal(t, data.PendingUpload.State, upload.Status, "upload status")
		}
	})
	t.Run("200 OK (Processing)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/"+data.ProcessingUpload.UUID.String(), nil)
		resp := test.DoRequest(h, r)

		var upload schema.Upload
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&upload), "decode upload") {
			assert.Equal(t, data.ProcessingUpload.UUID, upload.UUID, "upload UUID")
			assert.Equal(t, data.ProcessingUpload.State, upload.Status, "upload status")
			assert.Equal(t, data.ProcessingUpload.SongsTotal, upload.SongsTotal, "songs total")
			assert.Equal(t, data.ProcessingUpload.SongsProcessed, upload.SongsProcessed, "songs processed")
			assert.Equal(t, data.ProcessingUpload.Errors, upload.Errors, "errors")
		}
	})
	t.Run("200 OK (Done)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/"+data.UploadWithErrors.UUID.String(), nil)
		resp := test.DoRequest(h, r)

		var upload schema.Upload
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&upload), "decode upload") {
			assert.Equal(t, data.UploadWithErrors.UUID, upload.UUID, "upload UUID")
			assert.Equal(t, data.UploadWithErrors.State, upload.Status, "upload status")
			assert.Equal(t, data.UploadWithErrors.SongsTotal, upload.SongsTotal, "songs total")
			assert.Equal(t, data.UploadWithErrors.Errors, upload.Errors, "errors")
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodGet, "/"+data.AbsentUploadUUID.String(), http.StatusNotFound))
}

func TestController_Delete(t *testing.T) {
	h, _, data := setup(t, true)
	path := "/" + data.OpenUpload.UUID.String()

	t.Run("204 No Content", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, path, nil)
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Repeat the same delete to test idempotency
		r = httptest.NewRequest(http.MethodDelete, path, nil)
		resp = test.DoRequest(h, r)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID))
}
