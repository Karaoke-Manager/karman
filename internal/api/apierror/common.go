package apierror

import (
	"errors"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"gorm.io/gorm"
)

const (
	ProblemTypeDomain = "http://localhost/problems"
)

func BindError(err error) error {
	switch {
	case errors.Is(err, render.ErrUnsupportedFormat):
		return ErrUnsupportedMediaType
	default:
		return ErrUnprocessableEntity
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
