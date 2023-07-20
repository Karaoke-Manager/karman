package apierror

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"gorm.io/gorm"
)

const (
	ProblemTypeDomain   = "https://codello.dev/karman/problems"
	TypeValidationError = ProblemTypeDomain + "/validation-error"
)

func BindError(err error) *ProblemDetails {
	switch {
	case errors.Is(err, render.ErrUnsupportedFormat):
		return ErrUnsupportedMediaType
	case errors.As(err, &render.DecodeError{}):
		uerr := &json.UnmarshalTypeError{}
		if errors.As(err, &uerr) {
			return JSONUnmarshalError(uerr)
		}
		// Probably a syntax error
		return ErrBadRequest
	case errors.As(err, &render.BindError{}):
		return ErrUnprocessableEntity
	default:
		// Should not happen
		return ErrUnprocessableEntity
	}
}

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

func DBError(err error) *ProblemDetails {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	default:
		return ErrInternalServerError
	}
}
