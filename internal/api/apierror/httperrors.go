package apierror

import (
	"net/http"
	"strings"
)

const (
	// TypeMissingContentType indicates that a Content-Type header is required but was not specified.
	TypeMissingContentType = ProblemTypeDomain + "/missing-content-type"
)

// These errors are ProblemDetails representations of common HTTP error codes.
// These values do not have additional information associated with them and
// should only be used if the HTTP status code is sufficiently clear about the error.
var (
	ErrNotFound             = HttpStatus(http.StatusNotFound)
	ErrMethodNotAllowed     = HttpStatus(http.StatusMethodNotAllowed)
	ErrUnsupportedMediaType = HttpStatus(http.StatusUnsupportedMediaType)
	ErrInternalServerError  = HttpStatus(http.StatusInternalServerError)
	ErrUnprocessableEntity  = HttpStatus(http.StatusUnprocessableEntity)
	ErrBadRequest           = HttpStatus(http.StatusBadRequest)
)

// MissingContentType generates an error indicating that no content type was specified in the request.
// The allowed content types for this endpoint are included as an extra field.
func MissingContentType(allowed ...string) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeMissingContentType,
		Title:  "Missing Content-Type Header",
		Status: http.StatusBadRequest,
		Detail: "The HTTP header Content-Type is required but was not specified. " +
			"The following content types are allowed: " + strings.Join(allowed, ", "),
		Fields: map[string]any{
			"acceptedContentTypes": allowed,
		},
	}
}
