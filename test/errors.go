package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/Karaoke-Manager/karman/api/apierror"
)

// AssertProblemDetails validates that resp encodes a problem details instance with the specified values.
// Any fields will be checked for presence in the custom fields of the response.
// This assertion will NOT fail if additional fields are present in the response.
func AssertProblemDetails(t *testing.T, resp *http.Response, code int, errType string, fields map[string]any) {
	if resp.StatusCode != code {
		t.Errorf("%s %s responded with status code %d, expected %d", resp.Request.Method, resp.Request.RequestURI, resp.StatusCode, code)
	}
	if resp.Header.Get("Content-Type") != "application/problem+json" {
		t.Errorf("%s %s responded with Content-Type %s, expected %s", resp.Request.Method, resp.Request.RequestURI, resp.Header.Get("Content-Type"), "application/problem+json")
	}
	var details apierror.ProblemDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		t.Errorf("%s %s responded with invalid problem details schema: %s", resp.Request.Method, resp.Request.RequestURI, err)
		return
	}
	if details.Status != code {
		t.Errorf(`%s %s responded with {"code": %d}, expected %d`, resp.Request.Method, resp.Request.RequestURI, details.Status, code)
	}
	if details.IsDefaultType() && errType != "" {
		t.Errorf(`%s %s responded with {"type": <default>}, expected %q`, resp.Request.Method, resp.Request.RequestURI, errType)
	} else if !details.IsDefaultType() && errType == "" {
		t.Errorf(`%s %s responded with {"type": %q}, expected default type`, resp.Request.Method, resp.Request.RequestURI, details.Type)
	}
	for field, expected := range fields {
		actual, ok := details.Fields[field]
		if !ok {
			t.Errorf("%s %s responded with problem details, expected field %s", resp.Request.Method, resp.Request.RequestURI, field)
			continue
		}
		expectedType := reflect.TypeOf(expected)
		actualType := reflect.TypeOf(actual)
		if actualType.ConvertibleTo(expectedType) {
			actual = reflect.ValueOf(actual).Convert(expectedType).Interface()
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`%s %s responded with {%q: %q}, expected %q`, resp.Request.Method, resp.Request.RequestURI, field, actual, expected)
		}
	}
}

// AssertValidationError validates that resp encodes a problem details instance, indicating a 422 Unprocessable Entity error.
// It is validated that the error contains the specified error messages.
// errors maps from expected JSON pointers to their error messages.
func AssertValidationError(t *testing.T, resp *http.Response, errors map[string]string) {
	if errors == nil {
		AssertProblemDetails(t, resp, http.StatusUnprocessableEntity, apierror.TypeValidationError, nil)
	} else {
		expectedErrors := make([]any, 0, len(errors))
		for pointer, message := range errors {
			expectedErrors = append(expectedErrors, map[string]any{
				"pointer": pointer,
				"message": message,
			})
		}
		AssertProblemDetails(t, resp, http.StatusUnprocessableEntity, apierror.TypeValidationError, map[string]any{
			"errors": expectedErrors,
		})
	}
}

// MissingContentType returns a test that runs a request against h without the Content-Type header
// and validates that the response indicates as much.
// The variadic argument lets you specify the expected allowed content types for this endpoint.
func MissingContentType(h http.Handler, method string, path string, allowed ...any) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(method, path, nil)
		resp := DoRequest(h, r) //nolint:bodyclose
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
		resp := DoRequest(h, r) //nolint:bodyclose
		AssertProblemDetails(t, resp, http.StatusUnsupportedMediaType, apierror.TypeUnsupportedMediaType, map[string]any{
			"acceptedContentTypes": allowed,
		})
	}
}
