package test

import (
	"encoding/json"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// AssertProblemDetails validates that resp encodes a problem details instance with the specified values.
// Any fields will be checked for presence in the custom fields of the response.
// This assertion will NOT fail if additional fields are present in the response.
func AssertProblemDetails(t *testing.T, resp *http.Response, code int, errType string, fields map[string]any) {
	assert.Equal(t, code, resp.StatusCode, "response status code does not equal expected value")
	var err apierror.ProblemDetails
	if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&err), "response does not fit ProblemDetails schema") {
		assert.Equal(t, code, err.Status, "problem details status does not equal expected value")
		if errType == "" {
			assert.Truef(t, err.IsDefaultType(), "problem details has unexpected type %s", err.Type)
		} else {
			assert.Equal(t, errType, err.Type, "problem details has unexpected type")
		}
		for field, value := range fields {
			actual, ok := err.Fields[field]
			assert.Truef(t, ok, "field %s is not present in problem details", field)

			expectedType := reflect.TypeOf(value)
			actualType := reflect.TypeOf(actual)
			if actualType.ConvertibleTo(expectedType) {
				actual = reflect.ValueOf(actual).Convert(expectedType).Interface()
			}
			assert.Equalf(t, value, actual, "field %s has unexpected value in problem details", field)
		}
	}
}

// MissingContentType returns a test that runs a request against h without the Content-Type header
// and validates that the response indicates as much.
// The variadic argument lets you specify the expected allowed content types for this endpoint.
func MissingContentType(h http.Handler, method string, path string, allowed ...any) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		resp := w.Result()
		AssertProblemDetails(t, resp, http.StatusBadRequest, apierror.TypeMissingContentType, map[string]any{
			"acceptedContentTypes": allowed,
		})
	}
}

// InvalidContentType returns a test that runs a request against h with the specified invalid Content-Type header
// and validates that the response indicates as much.
// The variadic argument lets you specify the expected allowed content types for this endpoint.
func InvalidContentType(h http.Handler, method string, path string, invalid string, allowed ...any) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, path, nil)
		r.Header.Set("Content-Type", invalid)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		resp := w.Result()
		AssertProblemDetails(t, resp, http.StatusUnsupportedMediaType, apierror.TypeUnsupportedMediaType, map[string]any{
			"acceptedContentTypes": allowed,
		})
	}
}
