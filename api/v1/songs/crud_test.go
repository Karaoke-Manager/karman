//go:build database

package songs

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

//go:generate go run ../../../../tools/gensong -output testdata/valid-song.txt
func TestController_Create(t *testing.T) {
	t.Parallel()
	c, _ := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	url := "/v1/songs/"

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, url, test.MustOpen(t, "testdata/valid-song.txt"))
		r.Header.Set("Content-Type", "text/plain")
		resp := test.DoRequest(h, r)

		var song schema.Song
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("POST %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusCreated)
		}
		if err := json.NewDecoder(resp.Body).Decode(&song); err != nil {
			t.Errorf("POST %s responded with invalid song schema: %s", url, err)
		}
		if song.UUID == uuid.Nil {
			t.Errorf(`POST %s responded with {"uuid": null}, expected non-nil UUID`, url)
		}
		if song.Title != "Nineteen Eighty-Four" {
			t.Errorf(`POST %s responded with {"title": %q}, expected %q`, url, song.Title, "Nineteen Eighty-Four")
		}
	})
	t.Run("400 Bad Request (Body)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, url, strings.NewReader("Foo"))
		r.Header.Set("Content-Type", "text/plain")
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusBadRequest, apierror.TypeInvalidTXT, map[string]any{
			"line": 1,
		})
	})
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPost, url, "text/plain", "text/x-ultrastar"))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPost, url, "application/json", "text/plain", "text/x-ultrastar"))
}

func TestController_Find(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	testdata.NSongs(t, db, 150)
	url := "/v1/songs/"

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r)

		var songs []schema.Song
		test.AssertPagination(t, resp, 0, 25, 25, 150)
		if err := json.NewDecoder(resp.Body).Decode(&songs); err != nil {
			t.Errorf("GET %s responded with invalid song list schema: %s", url, err)
			return
		}
		if len(songs) != 25 {
			t.Errorf("GET %s responded with %d songs, expected %d", url, len(songs), 25)
		}
	})
	t.Run("400 Bad Request (Pagination)", test.InvalidPagination(h, http.MethodGet, url))
}

func TestController_Get(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	simpleSong := testdata.SimpleSong(t, db)
	url := fmt.Sprintf("/v1/songs/%s", simpleSong.UUID)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r)

		var song schema.Song
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if err := json.NewDecoder(resp.Body).Decode(&song); err != nil {
			t.Errorf("GET %s responded with invalid song schema: %s", url, err)
			return
		}
		if song.Genre != simpleSong.Genre {
			t.Errorf(`GET %s responded with {"genre": %q}, expected %q`, url, song.Genre, simpleSong.Genre)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s", testdata.InvalidUUID)))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s", uuid.New()), http.StatusNotFound))
}

func TestController_Update(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	simpleSong := testdata.SimpleSong(t, db)
	songWithUpload := testdata.SongWithUpload(t, db)
	url := fmt.Sprintf("/v1/songs/%s", simpleSong.UUID)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPatch, url, strings.NewReader(`
			{"title": "Foobar"}
		`))
		r.Header.Set("Content-Type", "application/json")
		resp := test.DoRequest(h, r)
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("PATCH %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusNoContent)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s", testdata.InvalidUUID)))
	t.Run("400 Bad Request (Invalid Body)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPatch, url, strings.NewReader(`
			{"title": "Foo
		`))
		r.Header.Set("Content-Type", "application/json")
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusBadRequest, "", nil)
	})
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPatch, url, "application/json"))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPatch, url, "text/plain", "application/json"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPatch, fmt.Sprintf("/v1/songs/%s", uuid.New()), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPatch, "/v1/songs/%s", songWithUpload.UUID))
	t.Run("422 Unprocessable Entity", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPatch, url, strings.NewReader(`
			{"title": "Foobar", "medley": {"mode": "manual"}}
		`))
		r.Header.Set("Content-Type", "application/json")
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusUnprocessableEntity, "", nil)
	})
}

func TestController_Delete(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	simpleSong := testdata.SimpleSong(t, db)
	url := fmt.Sprintf("/v1/songs/%s", simpleSong.UUID)

	t.Run("204 No Content", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, url, nil)
		resp := test.DoRequest(h, r)
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("DELETE %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusNoContent)
		}

		// Repeat the same delete to test idempotency
		r = httptest.NewRequest(http.MethodDelete, url, nil)
		resp = test.DoRequest(h, r)
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("DELETE %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusNoContent)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s", testdata.InvalidUUID)))
}
