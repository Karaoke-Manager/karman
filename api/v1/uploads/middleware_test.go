//go:build database

package uploads

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func TestHandler_FetchUpload(t *testing.T) {
	t.Parallel()

	h, db := setupHandler(t, "")
	openUpload := testdata.OpenUpload(t, db)

	m := func(t *testing.T) http.Handler {
		return h.FetchUpload(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			_, ok := GetUpload(r.Context())
			if !ok {
				t.Errorf("FetchUpload() did not set an upload in the context, expected upload to be set")
			}
		}))
	}

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/uploads/%s", openUpload.UUID), nil)
		r = r.WithContext(middleware.SetUUID(r.Context(), openUpload.UUID))
		test.DoRequest(m(t), r) //nolint:bodyclose
	})
	t.Run("404 Not Found", func(t *testing.T) {
		id := uuid.New()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/uploads/%s", id), nil)
		r = r.WithContext(middleware.SetUUID(r.Context(), id))
		resp := test.DoRequest(m(t), r) //nolint:bodyclose
		test.AssertProblemDetails(t, resp, http.StatusNotFound, "", nil)
	})
}

func TestHandler_ValidateFilePath(t *testing.T) {
	t.Parallel()

	h := func(t *testing.T) http.Handler {
		return ValidateFilePath(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			_, ok := GetFilePath(r.Context())
			if !ok {
				t.Errorf("ValidateFilePath() did not set a file path in the context, expected path to be set")
			}
		}))
	}

	t.Run("OK", func(t *testing.T) {
		path := "abc/def.txt"
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/uploads/{uuid}/files/%s", path), nil)
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("*", "abc/def.txt")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
		test.DoRequest(h(t), r) //nolint:bodyclose
	})
	t.Run("400 Bad Request", func(t *testing.T) {
		path := "some/../invalid-path"
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/uploads/{uuid}/files/%s", path), nil)
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("*", path)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
		resp := test.DoRequest(h(t), r) //nolint:bodyclose
		test.AssertProblemDetails(t, resp, http.StatusBadRequest, apierror.TypeInvalidUploadPath, map[string]any{
			"path": path,
		})
	})
}

func TestHandler_UploadState(t *testing.T) {
	t.Parallel()

	openUpload := model.Upload{
		Model: model.Model{UUID: uuid.New()},
		State: model.UploadStateOpen,
	}
	processingUpload := model.Upload{
		Model: model.Model{UUID: uuid.New()},
		State: model.UploadStateProcessing, SongsTotal: -1, SongsProcessed: -1,
	}

	h := UploadState(model.UploadStateOpen)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/uploads/%s/files/foo.txt", openUpload.UUID), nil)
		r = r.WithContext(SetUpload(r.Context(), openUpload))
		resp := test.DoRequest(h, r) //nolint:bodyclose
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("UploadState(%q) responded with status code %d, expected %d", model.UploadStateOpen, resp.StatusCode, http.StatusNoContent)
		}
	})
	t.Run("409 Conflict", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/uploads/%s/files/foo.txt", processingUpload.UUID), nil)
		r = r.WithContext(SetUpload(r.Context(), processingUpload))
		resp := test.DoRequest(h, r) //nolint:bodyclose
		test.AssertProblemDetails(t, resp, http.StatusConflict, apierror.TypeUploadState, map[string]any{
			"uuid": processingUpload.UUID.String(),
		})
	})
}
