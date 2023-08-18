package uploads

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/test"
)

func TestController_PutFile(t *testing.T) {
	h, _, data := setup(t, true)
	path := uploadPath(data.OpenUpload, "/files/foobar.txt")

	t.Run("204 No Content", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, path, strings.NewReader("Hello World"))
		r.Header.Set("Content-Type", "application/octet-stream")
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, "/"+data.InvalidUUID+"/files/foobar.txt"))
	t.Run("400 Bad Request (Bad Path)", testInvalidPath(h, http.MethodPut, uploadPath(data.OpenUpload, "/files/foo/../bar.txt"), "foo/../bar.txt"))
	t.Run("400 Bad Request (Root)", testInvalidPath(h, http.MethodPut, uploadPath(data.OpenUpload, "/files"), "."))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, path, "application/octet-stream"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, uploadPath(data.AbsentUpload, "/files/anything.txt"), http.StatusNotFound))
	t.Run("409 Conflict", testInvalidState(h, http.MethodPut, uploadPath(data.ProcessingUpload, "/files/test.txt"), data.ProcessingUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, path, "video/mp4", "application/octet-stream"))
}

func TestController_GetFile(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK (File)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, uploadPath(data.OpenUpload, "/files/foo/bar.txt"), nil)
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var stat schema.UploadFileStat
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&stat)) {
			assert.Equal(t, "bar.txt", stat.Name)
			assert.False(t, stat.Dir)
			assert.Nil(t, stat.Children)
		}
	})
	t.Run("200 OK (Folder)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, uploadPath(data.OpenUpload, "/files/foo"), nil)
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var stat schema.UploadFileStat
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&stat)) {
			assert.Equal(t, "foo", stat.Name)
			assert.True(t, stat.Dir)
			assert.Len(t, stat.Children, 1)
		}
	})
	t.Run("200 OK (Root)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, uploadPath(data.OpenUpload, "/files/"), nil)
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var stat schema.UploadFileStat
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&stat)) {
			assert.Equal(t, "", stat.Name)
			assert.True(t, stat.Dir)
			assert.Len(t, stat.Children, 2)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/files/abc.txt"))
	t.Run("400 Bad Request (Bad Path)", testInvalidPath(h, http.MethodGet, uploadPath(data.OpenUpload, "/files/foo/../bar.txt"), "foo/../bar.txt"))
	t.Run("404 Not Found (Upload)", test.HTTPError(h, http.MethodGet, uploadPath(data.AbsentUpload, "/files/anything.txt"), http.StatusNotFound))
	t.Run("404 Not Found (File)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, uploadPath(data.OpenUpload, "/files/absent.txt"), nil)
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusNotFound, apierror.TypeUploadFileNotFound, map[string]any{
			"uuid": data.OpenUpload.UUID.String(),
			"path": "absent.txt",
		})
	})
	t.Run("409 Conflict", testInvalidState(h, http.MethodGet, uploadPath(data.ProcessingUpload, "/files/anything.txt"), data.ProcessingUpload.UUID))
}

func TestController_DeleteFile(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("204 No Content", func(t *testing.T) {
		path := uploadPath(data.OpenUpload, "/files/foo/bar.txt")
		r := httptest.NewRequest(http.MethodDelete, path, nil)
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Repeat the request to test idempotency
		r = httptest.NewRequest(http.MethodDelete, path, nil)
		resp = test.DoRequest(h, r)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodDelete, "/"+data.InvalidUUID+"/files/abc.txt"))
	t.Run("400 Bad Request (Bad Path)", testInvalidPath(h, http.MethodDelete, uploadPath(data.OpenUpload, "/files/foo/../bar.txt"), "foo/../bar.txt"))
	t.Run("400 Bad Request (Root)", testInvalidPath(h, http.MethodDelete, uploadPath(data.OpenUpload, "/files"), "."))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodDelete, uploadPath(data.AbsentUpload, "/files/foobar.txt"), http.StatusNotFound))
	t.Run("409 Conflict", testInvalidState(h, http.MethodDelete, uploadPath(data.ProcessingUpload, "/files/anything.txt"), data.ProcessingUpload.UUID))
}
