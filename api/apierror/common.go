package apierror

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/Karaoke-Manager/karman/service/common"
)

// ProblemTypeDomain is the base domain for all custom problem types.
// This is the namespace under which all Karman problem types reside.
const ProblemTypeDomain = "https://codello.dev/karman/problems"

const (
	// TypeValidationError indicates a that the request data did not conform to the required schema.
	// This error should be associated with HTTP status code 422.
	TypeValidationError = ProblemTypeDomain + "/validation-error"

	// TypeInvalidUUID indicates that a UUID parameter was not a valid UUID.
	// This error should be associated with HTTP status code 400.
	TypeInvalidUUID = ProblemTypeDomain + "/invalid-uuid"
)

var (
	// ErrInvalidUUID is an error indicating that a UUID value was invalid.
	ErrInvalidUUID = &ProblemDetails{
		Type:   TypeInvalidUUID,
		Title:  "Invalid UUID",
		Status: http.StatusBadRequest,
	}
)

// BindError generates an error indicating that the request data was invalid in some way.
// This method should be used with errors generated from the render.Bind function.
func BindError(err error) *ProblemDetails {
	switch {
	case errors.Is(err, render.ErrMissingContentType):
		return BadRequest("The Content-Type header is required.")
	case errors.Is(err, render.ErrInvalidContentType):
		return BadRequest("The specified Content-Type is not correctly formatted.")
	case errors.Is(err, render.ErrNoMatchingDecoder):
		return ErrUnsupportedMediaType
	case errors.As(err, &render.DecodeError{}):
		uErr := &json.UnmarshalTypeError{}
		if errors.As(err, &uErr) {
			return JSONUnmarshalError(uErr)
		}
		// Probably a syntax error
		return ErrBadRequest
	case errors.As(err, &render.BindError{}):
		return UnprocessableEntity(errors.Unwrap(err).Error())
	default:
		// Should not happen
		return ErrUnprocessableEntity
	}
}

// JSONUnmarshalError generates an error indicating that the request data could not be parsed in some way.
func JSONUnmarshalError(err *json.UnmarshalTypeError) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeValidationError,
		Title:  "Validation Error",
		Status: 422,
		Detail: fmt.Sprintf("Expected type %s but got %s.", err.Type.Name(), err.Value),
		Fields: map[string]any{
			"field": err.Field,
		},
	}
}

// ServiceError generates an error indicating that a service request was not successful.
// This function maps known errors to their responses.
func ServiceError(err error) *ProblemDetails {
	switch {
	case errors.Is(err, common.ErrNotFound):
		return ErrNotFound
	default:
		return ErrInternalServerError
	}
}
