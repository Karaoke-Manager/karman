package apierror

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// ProblemTypeDomain is the base domain for all custom problem types.
// This is the namespace under which all Karman problem types reside.
const ProblemTypeDomain = "tag:codello.dev,2020:karman/problems:"

const (
	// TypeValidationError indicates a that the request data did not conform to the required schema.
	// This error should be associated with HTTP status code 422.
	TypeValidationError = ProblemTypeDomain + "validation-error"

	// TypeInvalidUUID indicates that a UUID parameter was not a valid UUID.
	// This error should be associated with HTTP status code 400.
	TypeInvalidUUID = ProblemTypeDomain + "invalid-uuid"
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
		// TODO: Support setting errors field when binding
		return ValidationError(errors.Unwrap(err).Error(), nil)
	default:
		// Should not happen
		return ErrUnprocessableEntity
	}
}

// JSONUnmarshalError generates an error indicating that the request data could not be parsed in some way.
func JSONUnmarshalError(err *json.UnmarshalTypeError) *ProblemDetails {
	field := "/" + strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(err.Field, "~", "~0"), "/", "~1"), ".", "/")
	return ValidationError(fmt.Sprintf("Expected type %s but got %s.", err.Type.Name(), err.Value), map[string]string{
		field: fmt.Sprintf("expected type %s, got %s", err.Type.Name(), err.Value),
	})
}

// ServiceError generates an error indicating that a service request was not successful.
// This function maps known errors to their responses.
func ServiceError(err error) *ProblemDetails {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return ErrNotFound
	default:
		return ErrInternalServerError
	}
}
