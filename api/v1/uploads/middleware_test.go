package uploads

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/test"
)

func TestController_FetchUpload(t *testing.T) {
	_, c, data := setup(t, true)
	h := c.FetchUpload(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUpload(r.Context())
		assert.True(t, ok, "Did not find an upload in the context.")
	}))

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(middleware.SetUUID(r.Context(), data.OpenUpload.UUID))
		test.DoRequest(h, r)
	})
	t.Run("404 Not Found", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(middleware.SetUUID(r.Context(), data.AbsentUploadUUID))
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusNotFound, "", nil)
	})
}

func TestController_ValidateFilePath(t *testing.T) {
	h := ValidateFilePath(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetFilePath(r.Context())
		assert.True(t, ok, "Did not find an upload in the context.")
	}))

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("*", "abc/def.txt")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
		test.DoRequest(h, r)
	})
	t.Run("400 Bad Request", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("*", "some//invalid path")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusBadRequest, apierror.TypeInvalidUploadPath, map[string]any{
			"path": "some//invalid path",
		})
	})
}

func TestController_UploadState(t *testing.T) {
	data := test.NewDataset(test.NewDB(t))
	h := UploadState(model.UploadStateOpen)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	t.Run("OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(SetUpload(r.Context(), data.OpenUpload))
		test.DoRequest(h, r)
	})
	t.Run("409 Conflict", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(SetUpload(r.Context(), data.ProcessingUpload))
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusConflict, apierror.TypeUploadState, map[string]any{
			"uuid": data.ProcessingUpload.UUID.String(),
		})
	})
}
