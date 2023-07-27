package apierror

import (
	"net/http"
	"strings"
)

const (
	// TypeMissingContentType indicates that a Content-Type header is required but was not specified.
	TypeMissingContentType = ProblemTypeDomain + "/missing-content-type"

	// TypeUnsupportedMediaType indicates that the Content-Type header was valid but the supplied media type is not allowed.
	TypeUnsupportedMediaType = ProblemTypeDomain + "/unsupported-media-type"
)

// These errors are ProblemDetails representations of common HTTP error codes.
// These values do not have additional information associated with them and
// should only be used if the HTTP status code by itself is sufficiently clear about the error.
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

// UnsupportedMediaType generates an error indicating that the Content-Type header contained an invalid value.
// The allowed content types for this endpoint are included as an extra field.
func UnsupportedMediaType(allowed ...string) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeUnsupportedMediaType,
		Status: http.StatusUnsupportedMediaType,
		Detail: "The following content types are allowed: " + strings.Join(allowed, ", "),
		Fields: map[string]any{
			"acceptedContentTypes": allowed,
		},
	}
}

// UnprocessableEntity generates an error indicating that the request payload did not conform to the expected schema.
func UnprocessableEntity(message string) *ProblemDetails {
	p := HttpStatus(http.StatusUnprocessableEntity)
	p.Detail = message
	return p
}
