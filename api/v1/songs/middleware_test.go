//go:build database

package songs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func TestController_FetchSong(t *testing.T) {
	t.Parallel()

	c, db := setupController(t)
	simpleSong := testdata.SimpleSong(t, db)

	h := c.FetchSong(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetSong(r.Context())
		assert.True(t, ok, "Did not find a song in the context.")
	}))

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(middleware.SetUUID(r.Context(), simpleSong.UUID))
		test.DoRequest(h, r)
	})
	t.Run("404 Not Found", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(middleware.SetUUID(r.Context(), uuid.New()))
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusNotFound, "", nil)
	})
}

func TestController_CheckModify(t *testing.T) {
	t.Parallel()

	c, _ := setupController(t)
	simpleSong := model.Song{Model: model.Model{UUID: uuid.New()}}
	songWithUpload := model.Song{Model: model.Model{UUID: uuid.New()}, InUpload: true}

	h := c.CheckModify(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(SetSong(r.Context(), simpleSong))
		test.DoRequest(h, r)
	})
	t.Run("409 Conflict", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(SetSong(r.Context(), songWithUpload))
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusConflict, apierror.TypeUploadSongReadonly, map[string]any{
			"uuid": songWithUpload.UUID.String(),
		})
	})
}
