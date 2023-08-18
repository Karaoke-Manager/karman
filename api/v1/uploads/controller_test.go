package uploads

import (
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	_ "github.com/Karaoke-Manager/karman/pkg/render/json"
	"github.com/Karaoke-Manager/karman/service/upload"
	"github.com/Karaoke-Manager/karman/test"
)

// setup prepares a test instance of the uploads.Controller.
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

	dir, err := os.MkdirTemp("", "karman-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	store, err := upload.NewFileStore(dir)
	require.NoError(t, err)

	svc := upload.NewService(db, store)
	c = NewController(svc)
	r := chi.NewRouter()
	r.Route("/", c.Router)
	return r, c, data
}
