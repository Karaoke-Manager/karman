package apierror

import (
	"net/http"
	"strings"
)

const (
	TypeMissingContentType = ProblemTypeDomain + "/missing-content-type"
)

var (
	ErrNotFound             = HttpStatus(http.StatusNotFound)
	ErrMethodNotAllowed     = HttpStatus(http.StatusMethodNotAllowed)
	ErrUnsupportedMediaType = HttpStatus(http.StatusUnsupportedMediaType)
	ErrInternalServerError  = HttpStatus(http.StatusInternalServerError)
	ErrUnprocessableEntity  = HttpStatus(http.StatusUnprocessableEntity)
	ErrBadRequest           = HttpStatus(http.StatusBadRequest)
)

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
