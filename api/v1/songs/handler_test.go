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
	"github.com/Karaoke-Manager/karman/pkg/nolog"
	_ "github.com/Karaoke-Manager/karman/pkg/render/json"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
	"github.com/Karaoke-Manager/karman/test"
)

// setupHandler prepares a test instance of Handler.
// The tests in this package are integration tests that run against an actual PostgreSQL database.
// The database can use testcontainers or be an external service.
func setupHandler(t *testing.T, prefix string) (*Handler, pgxutil.DB) {
	db := test.NewDB(t)
	songRepo := song.NewDBRepository(nolog.Logger, db)
	songSvc := song.NewService()
	mediaStore := media.NewMemStore()
	mediaRepo := media.NewDBRepository(nolog.Logger, db)
	mediaService := media.NewFakeService(mediaRepo)

	// workaround to support the prefix
	h := NewHandler(nolog.Logger, songRepo, songSvc, mediaStore, mediaService)
	r := chi.NewRouter()
	r.Mount(strings.TrimSuffix(prefix, "/")+"/", h.r)
	h.r = r
	return h, db
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
