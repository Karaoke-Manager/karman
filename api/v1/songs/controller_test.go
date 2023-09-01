//go:build database

package songs

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/api/apierror"
	_ "github.com/Karaoke-Manager/karman/pkg/render/json"
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
	songSvc := song.NewService()
	mediaStore := media.NewMemStore()
	mediaRepo := media.NewDBRepository(db)
	mediaService := media.NewFakeService(mediaRepo)
	return NewController(songRepo, songSvc, mediaStore, mediaService), db
}

// setupHandler is a convenience function that sets up a http.Handler for c.
func setupHandler(c *Controller, prefix string) http.Handler {
	r := chi.NewRouter()
	r.Route(strings.TrimSuffix(prefix, "/")+"/", c.Router)
	return r
}

// testSongConflict returns a test that checks that the specified request causes a 409 Conflict error.
func testSongConflict(h http.Handler, method string, urlFmt string, id uuid.UUID) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, fmt.Sprintf(urlFmt, id), nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose
		test.AssertProblemDetails(t, resp, http.StatusConflict, apierror.TypeUploadSongReadonly, map[string]any{
			"uuid": id.String(),
		})
	}
}
