package songs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/test"
)

//go:generate go run ../../../../tools/gensong -output testdata/valid-song.txt

func TestController_Create(t *testing.T) {
	h, _, _ := setup(t, false)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", test.MustOpen(t, "testdata/valid-song.txt"))
		r.Header.Set("Content-Type", "text/plain")
		resp := test.DoRequest(h, r)

		var song schema.Song
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&song), "response decode")
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "status Code")
		assert.NotEmpty(t, song.UUID)
		assert.Equal(t, "Nineteen Eighty-Four", song.Title, "song title")
	})
	t.Run("400 Bad Request (Body)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("Foo"))
		r.Header.Set("Content-Type", "text/plain")
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusBadRequest, apierror.TypeInvalidTXT, map[string]any{
			"line": 1,
		})
	})
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPost, "/", "text/plain", "text/x-ultrastar"))
	t.Run("400 Bad Request (Invalid Content-Type)", test.InvalidContentType(h, http.MethodPost, "/", "application/json", "text/plain", "text/x-ultrastar"))
}

func TestController_Find(t *testing.T) {
	h, _, _ := setup(t, true)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		resp := test.DoRequest(h, r)

		test.AssertPagination(t, resp, 0, 25, 25, 150)
		var songs []schema.Song
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&songs), "decode songs") {
			assert.Len(t, songs, 25, "length of result")
		}
	})
	t.Run("400 Bad Request (Pagination)", test.InvalidPagination(h, http.MethodGet, "/"))
}

func TestController_Get(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/"+data.BasicSong.UUID.String(), nil)
		resp := test.DoRequest(h, r)

		var song schema.Song
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&song), "decode song") {
			assert.Equal(t, schema.FromSong(data.BasicSong), song, "song data")
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodGet, "/"+data.AbsentSongUUID.String(), http.StatusNotFound))
}

func TestController_Update(t *testing.T) {
	h, _, data := setup(t, true)
	path := "/" + data.BasicSong.UUID.String()

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPatch, path, strings.NewReader(`
			{"title": "Foobar"}
		`))
		r.Header.Set("Content-Type", "application/json")
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID))
	t.Run("400 Bad Request (Invalid Body)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPatch, path, strings.NewReader(`
			{"title": "Foo
		`))
		r.Header.Set("Content-Type", "application/json")
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusBadRequest, "", nil)
	})
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPatch, path, "application/json"))
	t.Run("400 Bad Request (Invalid Content-Type)", test.InvalidContentType(h, http.MethodPatch, path, "text/plain", "application/json"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPatch, "/"+data.AbsentSongUUID.String(), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPatch, "/"+data.SongWithUpload.UUID.String(), data.SongWithUpload.UUID))
	t.Run("422 Unprocessable Entity", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPatch, path, strings.NewReader(`
			{"title": "Foobar", "medley": {"mode": "manual"}}
		`))
		r.Header.Set("Content-Type", "application/json")
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusUnprocessableEntity, "", nil)
	})
}

func TestController_Delete(t *testing.T) {
	h, _, data := setup(t, true)
	path := "/" + data.BasicSong.UUID.String()

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
