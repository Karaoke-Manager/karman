package apierror

import (
	"errors"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"gorm.io/gorm"
	"net/http"
)

func BindError(w http.ResponseWriter, r *http.Request, err error) error {
	if errors.Is(err, render.ErrUnsupportedFormat) {
		// TODO: Include the Accept header
		return render.Render(w, r, HttpStatus(http.StatusUnsupportedMediaType))
	}
	return render.Render(w, r, HttpStatus(http.StatusUnprocessableEntity))
}

func DBError(w http.ResponseWriter, r *http.Request, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound(w, r)
	} else {
		// TODO: Maybe there are other relevant errors we should differentiate
		// FIXME: Maybe status code Service Unavailable would be better here?
		return InternalServerError(w, r)
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) error {
	return render.Render(w, r, HttpStatus(http.StatusNotFound))
}

func InternalServerError(w http.ResponseWriter, r *http.Request) error {
	return render.Render(w, r, HttpStatus(http.StatusInternalServerError))
}
