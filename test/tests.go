package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Karaoke-Manager/karman/api/apierror"
)

// InvalidPagination returns a test that runs a request against h with invalid pagination request parameters
// and asserts an appropriate error response.
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

// APIError returns a test that runs a request against h and asserts that the response describes an error of type problemType.
func APIError(h http.Handler, method string, path string, status int, problemType string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, path, nil)
		resp := DoRequest(h, r)
		AssertProblemDetails(t, resp, status, problemType, nil)
	}
}

// HTTPError is a convenience function for APIError where we do not expect a non-default problem type.
func HTTPError(h http.Handler, method string, path string, status int) func(t *testing.T) {
	return APIError(h, method, path, status, "")
}

// InvalidUUID returns a test that runs a request against h and asserts that the response indicates an invalid UUID.
func InvalidUUID(h http.Handler, method string, path string) func(t *testing.T) {
	return APIError(h, method, path, http.StatusBadRequest, apierror.TypeInvalidUUID)
}
