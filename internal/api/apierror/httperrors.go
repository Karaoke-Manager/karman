package apierror

import (
	"net/http"
)

var (
	ErrNotFound             = HttpStatus(http.StatusNotFound)
	ErrMethodNotAllowed     = HttpStatus(http.StatusMethodNotAllowed)
	ErrUnsupportedMediaType = HttpStatus(http.StatusUnsupportedMediaType)
	ErrInternalServerError  = HttpStatus(http.StatusInternalServerError)
	ErrUnprocessableEntity  = HttpStatus(http.StatusUnprocessableEntity)
)
