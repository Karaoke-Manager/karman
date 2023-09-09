package apierror

import (
	"net/http"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

const (
	// TypeMissingContentType indicates that a Content-Type header is required but was not specified.
	TypeMissingContentType = ProblemTypeDomain + "missing-content-type"

	// TypeUnsupportedMediaType indicates that the Content-Type header was valid but the supplied media type is not allowed.
	TypeUnsupportedMediaType = ProblemTypeDomain + "unsupported-media-type"
)

// These errors are ProblemDetails representations of common HTTP error codes.
// These values do not have additional information associated with them and
// should only be used if the HTTP status code by itself is sufficiently clear about the error.
var (
	ErrBadRequest           = HTTPStatus(http.StatusBadRequest)
	ErrNotFound             = HTTPStatus(http.StatusNotFound)
	ErrMethodNotAllowed     = HTTPStatus(http.StatusMethodNotAllowed)
	ErrNotAcceptable        = HTTPStatus(http.StatusNotAcceptable)
	ErrUnprocessableEntity  = HTTPStatus(http.StatusUnprocessableEntity)
	ErrUnsupportedMediaType = HTTPStatus(http.StatusUnsupportedMediaType)
	ErrInternalServerError  = HTTPStatus(http.StatusInternalServerError)
	ErrServiceUnavailable   = HTTPStatus(http.StatusServiceUnavailable)
)

// MissingContentType generates an error indicating that no content type was specified in the request.
// The allowed content types for this endpoint are included as an extra field.
func MissingContentType(allowed ...mediatype.MediaType) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeMissingContentType,
		Title:  "Missing Content-Type Header",
		Status: http.StatusBadRequest,
		Detail: "The HTTP header Content-Type is required but was not specified.",
		Fields: map[string]any{
			"acceptedContentTypes": allowed,
		},
	}
}

// UnsupportedMediaType generates an error indicating that the Content-Type header contained an invalid value.
// The allowed content types for this endpoint are included as an extra field.
func UnsupportedMediaType(allowed ...mediatype.MediaType) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeUnsupportedMediaType,
		Status: http.StatusUnsupportedMediaType,
		Fields: map[string]any{
			"acceptedContentTypes": allowed,
		},
	}
}

// ValidationError generates an error indicating that the request payload did not conform to the expected schema.
func ValidationError(message string, errors map[string]string) *ProblemDetails {
	err := &ProblemDetails{
		Type:   TypeValidationError,
		Title:  "Unprocessable Entity",
		Status: http.StatusUnprocessableEntity,
		Detail: message,
	}
	if errors != nil {
		errorList := make([]map[string]string, 0, len(errors))
		for pointer, msg := range errors {
			errorList = append(errorList, map[string]string{
				"pointer": pointer,
				"message": msg,
			})
		}
		err.Fields = map[string]any{"errors": errorList}
	}
	return err
}

// BadRequest generates an 400 Bad Request error with the specified message.
func BadRequest(message string) *ProblemDetails {
	p := HTTPStatus(http.StatusBadRequest)
	p.Detail = message
	return p
}
