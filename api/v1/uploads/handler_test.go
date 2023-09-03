//go:build database

package uploads

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/api/apierror"
	_ "github.com/Karaoke-Manager/karman/pkg/render/json"
	"github.com/Karaoke-Manager/karman/service/upload"
	"github.com/Karaoke-Manager/karman/test"
)

// setupHandler prepares a test instance of Handler.
// The tests in this package are integration tests that run against an actual PostgreSQL database.
// The database can use testcontainers or be an external service.
func setupHandler(t *testing.T, prefix string) (*Handler, pgxutil.DB) {
	dir, err := os.MkdirTemp("", "karman-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp() returned an unexpected error: %s", err)
	}
	db := test.NewDB(t)
	uploadRepo := upload.NewDBRepository(db)
	uploadStore, err := upload.NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore(%q) returned an unexpected error: %s", dir, err)
	}

	// workaround to support the prefix
	h := NewHandler(uploadRepo, uploadStore)
	r := chi.NewRouter()
	r.Mount(strings.TrimSuffix(prefix, "/")+"/", h.r)
	h.r = r
	return h, db
}

// setupFiles creates files for testing in the store backing c.
// The files are created for an upload with UUID id.
// The files map maps filenames (or paths) to file contents.
// File paths must be valid according to fs.ValidPath.
func setupFiles(t *testing.T, h *Handler, id uuid.UUID, files map[string]string) {
	for file, content := range files {
		w, err := h.uploadStore.Create(context.TODO(), id, file)
		if err != nil {
			t.Fatalf("Create(ctx, %q, %q) returned an unexpected error: %s", id, file, err)
		}
		if _, err = io.WriteString(w, content); err != nil {
			t.Fatalf("w.WriteString(w, %q) returned an unexpected error: %s", content, err)
		}
		if err = w.Close(); err != nil {
			t.Fatalf("w.Close() returned an unexpected error: %s", err)
		}
	}
}

// testInvalidPath is a test, that performs a request using the specified method and request path.
// It then asserts that the response indicates an invalid file path that contains the path value.
func testInvalidPath(h http.Handler, method string, reqPath string, path string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, reqPath, nil)
		r.Header.Set("Content-Type", "application/octet-stream")
		resp := test.DoRequest(h, r) //nolint:bodyclose
		test.AssertProblemDetails(t, resp, http.StatusBadRequest, apierror.TypeInvalidUploadPath, map[string]any{
			"path": path,
		})
	}
}

// testInvalidState is a test that performs a request using the specified method and path.
// It then asserts that the response indicates an invalid upload state and contains the specified UUID.
func testInvalidState(h http.Handler, method string, pathFmt string, id uuid.UUID) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, fmt.Sprintf(pathFmt, id), nil)
		resp := test.DoRequest(h, r) //nolint:bodyclose
		test.AssertProblemDetails(t, resp, http.StatusConflict, apierror.TypeUploadState, map[string]any{
			"uuid": id.String(),
		})
	}
}
