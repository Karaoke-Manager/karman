//go:build database

package songs

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func TestHandler_FetchSong(t *testing.T) {
	t.Parallel()

	h, db := setupHandler(t, "")
	simpleSong := testdata.SimpleSong(t, db)

	m := h.FetchSong(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetSong(r.Context())
		if !ok {
			t.Errorf("FetchSong() did not set a song in the context, expected song to be set")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/songs/%s", simpleSong.UUID), nil)
		r = r.WithContext(middleware.SetUUID(r.Context(), simpleSong.UUID))
		resp := test.DoRequest(m, r) //nolint:bodyclose
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("FetchSong() responded with status code %d, expected %d", resp.StatusCode, http.StatusNoContent)
		}
	})
	t.Run("404 Not Found", func(t *testing.T) {
		id := uuid.New()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/songs/%s", id), nil)
		r = r.WithContext(middleware.SetUUID(r.Context(), id))
		resp := test.DoRequest(m, r) //nolint:bodyclose
		test.AssertProblemDetails(t, resp, http.StatusNotFound, "", nil)
	})
}

func TestHandler_CheckModify(t *testing.T) {
	t.Parallel()

	h, _ := setupHandler(t, "")
	simpleSong := model.Song{Model: model.Model{UUID: uuid.New()}}
	songWithUpload := model.Song{Model: model.Model{UUID: uuid.New()}, InUpload: true}

	m := h.CheckModify(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/songs/%s/txt", simpleSong.UUID), nil)
		r = r.WithContext(SetSong(r.Context(), simpleSong))
		resp := test.DoRequest(m, r) //nolint:bodyclose
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("CheckModify() responded with status code %d, expected %d", resp.StatusCode, http.StatusNoContent)
		}
	})
	t.Run("409 Conflict", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/songs/%s/txt", songWithUpload.UUID), nil)
		r = r.WithContext(SetSong(r.Context(), songWithUpload))
		resp := test.DoRequest(m, r) //nolint:bodyclose
		test.AssertProblemDetails(t, resp, http.StatusConflict, apierror.TypeUploadSongReadonly, map[string]any{
			"uuid": songWithUpload.UUID.String(),
		})
	})
}
