package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Karaoke-Manager/karman/internal/api/apierror"
)

func InvalidPagination(h http.Handler, method string, path string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, path, nil)
		q := r.URL.Query()
		q.Set("limit", "foo")
		q.Set("offset", "bar")
		r.URL.RawQuery = q.Encode()
		resp := DoRequest(h, r)
		AssertProblemDetails(t, resp, http.StatusBadRequest, "", nil)
	}
}

func APIError(h http.Handler, method string, path string, status int, problemType string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, path, nil)
		resp := DoRequest(h, r)
		AssertProblemDetails(t, resp, status, problemType, nil)
	}
}

func HTTPError(h http.Handler, method string, path string, status int) func(t *testing.T) {
	return APIError(h, method, path, status, "")
}

func InvalidUUID(h http.Handler, method string, path string) func(t *testing.T) {
	return APIError(h, method, path, http.StatusBadRequest, apierror.TypeInvalidUUID)
}
