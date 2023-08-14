package songs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Karaoke-Manager/server/internal/api/apierror"
	"github.com/Karaoke-Manager/server/internal/model"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/server/internal/service/media"
	"github.com/Karaoke-Manager/server/internal/service/song"
	"github.com/Karaoke-Manager/server/internal/test"
)

// setup prepares a test instance of the songs.Controller.
// The tests in this package are more integration tests than unit tests as we test against an in-memory SQLite database
// instead of mocking service objects.
// The reason for this approach is mainly reduced testing complexity.
//
// If withData is true, a test dataset will be created and stored in the DB.
// Otherwise, data will be nil.
func setup(t *testing.T, withData bool) (h http.Handler, c *Controller, data *test.Dataset) {
	db := test.NewDB(t)
	if withData {
		data = test.NewDataset(db)
	}
	songSvc := song.NewService(db)
	mediaSvc := media.NewFakeService("Foobar", db)
	c = NewController(songSvc, mediaSvc)
	r := chi.NewRouter()
	r.Route("/", c.Router)
	return r, c, data
}

func songPath(song *model.Song, suffix string) string {
	return "/" + song.UUID.String() + suffix
}

func testSongConflict(h http.Handler, method string, path string, id uuid.UUID) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, path, nil)
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusConflict, apierror.TypeUploadSongReadonly, map[string]any{
			"uuid": id.String(),
		})
	}
}
