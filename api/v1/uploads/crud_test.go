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
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func TestController_Create(t *testing.T) {
	t.Parallel()
	c, _ := setupController(t)
	h := setupHandler(c, "/v1/uploads/")
	url := "/v1/uploads/"

	t.Run("201 Created", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose

		var upload schema.Upload
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("POST %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusCreated)
		}
		if err := json.NewDecoder(resp.Body).Decode(&upload); err != nil {
			t.Errorf("POST %s responded with invalid upload schema: %s", url, err)
			return
		}
		if upload.UUID == uuid.Nil {
			t.Errorf("POST %s responded with no upload UUID, expected non-nil UUID", url)
		}
		if upload.Status != model.UploadStateOpen {
			t.Errorf(`POST %s responded with {"status": %q}, expected %q`, url, upload.Status, model.UploadStateOpen)
		}
	})
}

func TestController_Find(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/uploads/")
	testdata.NOpenUploads(t, db, 15)
	testdata.NPendingUploads(t, db, 15)
	url := "/v1/uploads/"

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose

		test.AssertPagination(t, resp, 0, 25, 25, 30)
		var uploads []schema.Upload
		if err := json.NewDecoder(resp.Body).Decode(&uploads); err != nil {
			t.Errorf("GET %s responded with invalid upload list schema: %s", url, err)
			return
		}
		if len(uploads) != 25 {
			t.Errorf("GET %s responded with %d uploads, expected %d", url, len(uploads), 25)
		}
	})
	t.Run("400 Bad Request (Pagination)", test.InvalidPagination(h, http.MethodGet, url))
}

func TestController_Get(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/uploads/")
	openUpload := testdata.OpenUpload(t, db)
	pendingUpload := testdata.PendingUpload(t, db)
	processingUpload := testdata.ProcessingUpload(t, db)
	doneUpload := testdata.DoneUpload(t, db)

	t.Run("200 OK (Open)", func(t *testing.T) {
		url := fmt.Sprintf("/v1/uploads/%s", openUpload.UUID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose

		var upload schema.Upload
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if err := json.NewDecoder(resp.Body).Decode(&upload); err != nil {
			t.Errorf("GET %s responded with invalid upload schema: %s", url, err)
			return
		}
		if upload.UUID != openUpload.UUID {
			t.Errorf(`GET %s responded with {"uuid": %q}, expected %q`, url, upload.UUID, openUpload.UUID)
		}
		if upload.Status != openUpload.State {
			t.Errorf(`GET %s responded with {"status": %q}, expected %q`, url, upload.Status, openUpload.State)
		}
	})
	t.Run("200 OK (Pending)", func(t *testing.T) {
		url := fmt.Sprintf("/v1/uploads/%s", pendingUpload.UUID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose

		var upload schema.Upload
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if err := json.NewDecoder(resp.Body).Decode(&upload); err != nil {
			t.Errorf("GET %s responded with invalid upload schema: %s", url, err)
			return
		}
		if upload.UUID != pendingUpload.UUID {
			t.Errorf(`GET %s responded with {"uuid": %q}, expected %q`, url, upload.UUID, pendingUpload.UUID)
		}
		if upload.Status != pendingUpload.State {
			t.Errorf(`GET %s responded with {"status": %q}, expected %q`, url, upload.Status, pendingUpload.State)
		}
	})
	t.Run("200 OK (Processing)", func(t *testing.T) {
		url := fmt.Sprintf("/v1/uploads/%s", processingUpload.UUID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose

		var upload schema.Upload
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if err := json.NewDecoder(resp.Body).Decode(&upload); err != nil {
			t.Errorf("GET %s responded with invalid upload schema: %s", url, err)
			return
		}
		if upload.UUID != processingUpload.UUID {
			t.Errorf(`GET %s responded with {"uuid": %q}, expected %q`, url, upload.UUID, processingUpload.UUID)
		}
		if upload.Status != processingUpload.State {
			t.Errorf(`GET %s responded with {"status": %q}, expected %q`, url, upload.Status, processingUpload.State)
		}
		if upload.SongsTotal != processingUpload.SongsTotal {
			t.Errorf(`GET %s responded with {"songsTotal": %d}, expected %d`, url, upload.SongsTotal, processingUpload.SongsTotal)
		}
		if upload.SongsProcessed != processingUpload.SongsProcessed {
			t.Errorf(`GET %s responded with {"songsProcessed": %d}, expected %d`, url, upload.SongsProcessed, processingUpload.SongsProcessed)
		}
		if upload.Errors != processingUpload.Errors {
			t.Errorf(`GET %s responded with {"errors": %d}, expected %d`, url, upload.Errors, processingUpload.Errors)
		}
	})
	t.Run("200 OK (Done)", func(t *testing.T) {
		url := fmt.Sprintf("/v1/uploads/%s", doneUpload.UUID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose

		var upload schema.Upload
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if err := json.NewDecoder(resp.Body).Decode(&upload); err != nil {
			t.Errorf("GET %s responded with invalid upload schema: %s", url, err)
			return
		}
		if upload.UUID != doneUpload.UUID {
			t.Errorf(`GET %s responded with {"uuid": %q}, expected %q`, url, upload.UUID, doneUpload.UUID)
		}
		if upload.Status != doneUpload.State {
			t.Errorf(`GET %s responded with {"status": %q}, expected %q`, url, upload.Status, doneUpload.State)
		}
		if upload.SongsTotal != doneUpload.SongsTotal {
			t.Errorf(`GET %s responded with {"songsTotal": %d}, expected %d`, url, upload.SongsTotal, doneUpload.SongsTotal)
		}
		if upload.Errors != doneUpload.Errors {
			t.Errorf(`GET %s responded with {"errors": %d}, expected %d`, url, upload.Errors, doneUpload.Errors)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/uploads/%s", testdata.InvalidUUID)))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodGet, fmt.Sprintf("/v1/uploads/%s", uuid.New()), http.StatusNotFound))
}

func TestController_Delete(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/uploads/")
	openUpload := testdata.OpenUpload(t, db)
	url := fmt.Sprintf("/v1/uploads/%s", openUpload.UUID)

	t.Run("204 No Content", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, url, nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("DLEETE %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusNoContent)
		}

		// Repeat the same delete to test idempotency
		r = httptest.NewRequest(http.MethodDelete, url, nil)
		resp = test.DoRequest(h, r) //nolint:bodyclose
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("DLEETE %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusNoContent)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/uploads/%s", testdata.InvalidUUID)))
}
