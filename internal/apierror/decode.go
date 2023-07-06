package apierror

import (
	"errors"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"gorm.io/gorm"
	"net/http"
)

func DecodeError(w http.ResponseWriter, r *http.Request, err error) error {
	if errors.Is(err, render.ErrUnsupportedFormat) {
		// TODO: Include the Accept header
		return render.Render(w, r, HttpStatus(http.StatusUnsupportedMediaType))
	}
	return render.Render(w, r, HttpStatus(http.StatusUnprocessableEntity))
}

func DBError(w http.ResponseWriter, r *http.Request, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return render.Render(w, r, HttpStatus(http.StatusNotFound))
	} else {
		// TODO: Maybe there are other relevant errors we should differentiate
		// FIXME: Maybe status code Service Unavailable would be better here?
		return render.Render(w, r, HttpStatus(http.StatusInternalServerError))
	}
}
