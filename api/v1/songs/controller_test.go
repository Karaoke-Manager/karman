//go:build database

package songs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
	"github.com/Karaoke-Manager/karman/test"
)

// setupController prepares a test instance of the songs.Controller.
// The tests in this package are integration tests that run against an actual PostgreSQL database.
// The database can use testcontainers or be an external service.
func setupController(t *testing.T) (*Controller, pgxutil.DB) {
	db := test.NewDB(t)
	songRepo := song.NewDBRepository(db)
	mediaStore := media.NewFakeStore()
	mediaService := media.NewFakeService(media.NewDBRepository(db))
	return NewController(songRepo, mediaStore, mediaService), db
}

// setupHandler is a convenience function that calls setupController and builds a HTTP handler around the controller.
func setupHandler(t *testing.T, prefix string) (http.Handler, pgxutil.DB) {
	c, db := setupController(t)
	r := chi.NewRouter()
	r.Route(strings.TrimSuffix(prefix, "/")+"/", c.Router)
	return r, db
}

// songPath is a helper function that returns the request path to the resource identified by suffix, scoped to song.
func songPath(song model.Song, suffix string) string {
	return "/" + song.UUID.String() + suffix
}

// testSongConflict returns a test that checks that the specified request causes a 409 Conflict error.
func testSongConflict(h http.Handler, method string, path string, id uuid.UUID) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, path, nil)
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusConflict, apierror.TypeUploadSongReadonly, map[string]any{
			"uuid": id.String(),
		})
	}
}
