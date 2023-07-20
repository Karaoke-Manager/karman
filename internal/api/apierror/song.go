package apierror

import (
	"errors"
	"github.com/Karaoke-Manager/go-ultrastar/txt"
	"net/http"
)

const (
	TypeInvalidTXT = ProblemTypeDomain + "/invalid-ultrastar-txt"
)

func InvalidUltraStarTXT(err error) *ProblemDetails {
	var parseErr txt.ParseError
	var fields map[string]any
	if errors.As(err, &parseErr) {
		fields = map[string]any{
			"line": parseErr.Line(),
		}
	}
	return &ProblemDetails{
		Type:   TypeInvalidTXT,
		Title:  "Invalid UltraStar TXT format",
		Status: http.StatusBadRequest,
		Detail: err.Error(),
		Fields: fields,
	}
}
