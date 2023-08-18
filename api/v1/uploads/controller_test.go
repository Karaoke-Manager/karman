package uploads

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/model"
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
	if withData {
		ctx := context.Background()
		w, err := svc.CreateFile(ctx, data.OpenUpload, "foo/bar.txt")
		require.NoError(t, err)
		_, err = io.WriteString(w, "Hello World")
		require.NoError(t, err)
		require.NoError(t, w.Close())
		w, err = svc.CreateFile(ctx, data.OpenUpload, "test.txt")
		require.NoError(t, err)
		_, err = io.WriteString(w, "Foobar")
		require.NoError(t, err)
		require.NoError(t, w.Close())
	}

	c = NewController(svc)
	r := chi.NewRouter()
	r.Route("/", c.Router)
	return r, c, data
}

func uploadPath(upload *model.Upload, suffix string) string {
	return "/" + upload.UUID.String() + suffix
}

func testInvalidPath(h http.Handler, method string, reqPath string, path string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, reqPath, nil)
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusBadRequest, apierror.TypeInvalidUploadPath, map[string]any{
			"path": path,
		})
	}
}

func testInvalidState(h http.Handler, method string, path string, id uuid.UUID) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, path, nil)
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusConflict, apierror.TypeUploadState, map[string]any{
			"uuid": id.String(),
		})
	}
}
