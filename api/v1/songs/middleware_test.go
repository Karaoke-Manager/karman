package songs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/test"
)

func TestController_FetchUpload(t *testing.T) {
	_, c, data := setup(t, true)
	h := c.FetchUpload(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetSong(r.Context())
		assert.True(t, ok, "Did not find a song in the context.")
	}))

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(middleware.SetUUID(r.Context(), data.BasicSong.UUID))
		test.DoRequest(h, r)
	})
	t.Run("404 Not Found", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(middleware.SetUUID(r.Context(), data.AbsentSongUUID))
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusNotFound, "", nil)
	})
}

func TestController_CheckModify(t *testing.T) {
	_, c, data := setup(t, true)
	h := c.CheckModify(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(SetSong(r.Context(), data.BasicSong))
		test.DoRequest(h, r)
	})
	t.Run("409 Conflict", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(SetSong(r.Context(), data.SongWithUpload))
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusConflict, apierror.TypeUploadSongReadonly, map[string]any{
			"uuid": data.SongWithUpload.UUID.String(),
		})
	})
}
