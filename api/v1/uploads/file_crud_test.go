//go:build database

package uploads

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func TestHandler_PutFile(t *testing.T) {
	t.Parallel()

	h, db := setupHandler(t, "/v1/uploads/")
	openUpload := testdata.OpenUpload(t, db)
	processingUpload := testdata.ProcessingUpload(t, db)
	url := fmt.Sprintf("/v1/uploads/%s/files/foobar.txt", openUpload.UUID)

	t.Run("204 No Content", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, url, strings.NewReader("Hello World"))
		r.Header.Set("Content-Type", "application/octet-stream")
		resp := test.DoRequest(h, r) //nolint:bodyclose
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("PUT %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusNoContent)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, fmt.Sprintf("/v1/uploads/%s/files/foobar.txt", testdata.InvalidUUID)))
	t.Run("400 Bad Request (Bad Path)", testInvalidPath(h, http.MethodPut, fmt.Sprintf("/v1/uploads/%s/files/foo/../bar.txt", openUpload.UUID), "foo/../bar.txt"))
	t.Run("400 Bad Request (Root)", testInvalidPath(h, http.MethodPut, fmt.Sprintf("/v1/uploads/%s/files", openUpload.UUID), "."))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, url, "application/octet-stream"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, fmt.Sprintf("/v1/uploads/%s/files/foobar.txt", uuid.New()), http.StatusNotFound))
	t.Run("409 Conflict", testInvalidState(h, http.MethodPut, "/v1/uploads/%s/files/foobar.txt", processingUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, url, "video/mp4", "application/octet-stream"))
}

func TestHandler_GetFile(t *testing.T) {
	t.Parallel()

	h, db := setupHandler(t, "/v1/uploads/")
	openUpload := testdata.OpenUpload(t, db)
	processingUpload := testdata.ProcessingUpload(t, db)
	setupFiles(t, h, openUpload.UUID, map[string]string{
		"foo/bar.txt": "Hello World",
		"test.txt":    "Nothing",
	})

	t.Run("200 OK (File)", func(t *testing.T) {
		url := fmt.Sprintf("/v1/uploads/%s/files/foo/bar.txt", openUpload.UUID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose
		var stat schema.UploadFileStat
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if err := json.NewDecoder(resp.Body).Decode(&stat); err != nil {
			t.Errorf("GET %s responded with invalid upload file schema: %s", url, err)
			return
		}
		if stat.Name != "bar.txt" {
			t.Errorf(`GET %s responded with {"name": %q}, expected %q`, url, stat.Name, "bar.txt")
		}
		if stat.Dir {
			t.Errorf(`GET %s responded with {"dir": true}, expected false`, url)
		}
		if stat.Children != nil {
			t.Errorf(`GET %s responded with {"children": %v}, expected none`, url, stat.Children)
		}
	})
	t.Run("200 OK (Folder)", func(t *testing.T) {
		url := fmt.Sprintf("/v1/uploads/%s/files/foo", openUpload.UUID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose
		var stat schema.UploadFileStat
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if err := json.NewDecoder(resp.Body).Decode(&stat); err != nil {
			t.Errorf("GET %s responded with invalid upload file schema: %s", url, err)
			return
		}
		if stat.Name != "foo" {
			t.Errorf(`GET %s responded with {"name": %q}, expected %q`, url, stat.Name, "foo")
		}
		if !stat.Dir {
			t.Errorf(`GET %s responded with {"dir": false}, expected true`, url)
		}
		if len(stat.Children) != 1 {
			t.Errorf(`GET %s responded with %d children, expected %d`, url, len(stat.Children), 1)
		}
	})
	t.Run("200 OK (Root)", func(t *testing.T) {
		url := fmt.Sprintf("/v1/uploads/%s/files/", openUpload.UUID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose
		var stat schema.UploadFileStat
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if err := json.NewDecoder(resp.Body).Decode(&stat); err != nil {
			t.Errorf("GET %s responded with invalid upload file schema: %s", url, err)
			return
		}
		if stat.Name != "" {
			t.Errorf(`GET %s responded with {"name": %q}, expected %q`, url, stat.Name, "")
		}
		if !stat.Dir {
			t.Errorf(`GET %s responded with {"dir": false}, expected true`, url)
		}
		if len(stat.Children) != 2 {
			t.Errorf(`GET %s responded with %d children, expected %d`, url, len(stat.Children), 2)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/uploads/%s/files/foo/bar.txt", testdata.InvalidUUID)))
	t.Run("400 Bad Request (Bad Path)", testInvalidPath(h, http.MethodGet, fmt.Sprintf("/v1/uploads/%s/files/foo/../bar.txt", openUpload.UUID), "foo/../bar.txt"))
	t.Run("404 Not Found (Upload)", test.HTTPError(h, http.MethodGet, fmt.Sprintf("/v1/uploads/%s/files/test.txt", uuid.New()), http.StatusNotFound))
	t.Run("404 Not Found (File)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/uploads/%s/files/absent.txt", openUpload.UUID), nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose
		test.AssertProblemDetails(t, resp, http.StatusNotFound, apierror.TypeUploadFileNotFound, map[string]any{
			"uuid": openUpload.UUID.String(),
			"path": "absent.txt",
		})
	})
	t.Run("409 Conflict", testInvalidState(h, http.MethodGet, "/v1/uploads/%s/files/test.txt", processingUpload.UUID))
}

func TestHandler_DeleteFile(t *testing.T) {
	t.Parallel()

	h, db := setupHandler(t, "/v1/uploads/")
	openUpload := testdata.OpenUpload(t, db)
	processingUpload := testdata.ProcessingUpload(t, db)
	setupFiles(t, h, openUpload.UUID, map[string]string{
		"foo/bar.txt": "Hello World",
		"test.txt":    "Nothing",
	})

	t.Run("204 No Content (File)", func(t *testing.T) {
		url := fmt.Sprintf("/v1/uploads/%s/files/test.txt", openUpload.UUID)
		r := httptest.NewRequest(http.MethodDelete, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("DELETE %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusNoContent)
		}

		// Repeat the request to test idempotency
		r = httptest.NewRequest(http.MethodDelete, url, nil)
		resp = test.DoRequest(h, r) //nolint:bodyclose
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("DELETE %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusNoContent)
		}
	})
	t.Run("204 No Content (Folder)", func(t *testing.T) {
		url := fmt.Sprintf("/v1/uploads/%s/files/foo", openUpload.UUID)
		r := httptest.NewRequest(http.MethodDelete, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("DELETE %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusNoContent)
		}

		// Repeat the request to test idempotency
		r = httptest.NewRequest(http.MethodDelete, url, nil)
		resp = test.DoRequest(h, r) //nolint:bodyclose
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("DELETE %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusNoContent)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodDelete, fmt.Sprintf("/v1/uploads/%s/files/test.txt", testdata.InvalidUUID)))
	t.Run("400 Bad Request (Bad Path)", testInvalidPath(h, http.MethodDelete, fmt.Sprintf("/v1/uploads/%s/files/foo/../bar.txt", openUpload.UUID), "foo/../bar.txt"))
	t.Run("400 Bad Request (Root)", testInvalidPath(h, http.MethodDelete, fmt.Sprintf("/v1/uploads/%s/files", openUpload.UUID), "."))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodDelete, fmt.Sprintf("/v1/uploads/%s/files/foobar.txt", uuid.New()), http.StatusNotFound))
	t.Run("409 Conflict", testInvalidState(h, http.MethodDelete, "/v1/uploads/%s/files/anything.txt", processingUpload.UUID))
}
