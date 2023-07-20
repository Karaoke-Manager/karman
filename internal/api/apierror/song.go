package apierror

import (
	"errors"
	"github.com/Karaoke-Manager/go-ultrastar/txt"
	"net/http"
)

const (
	// TypeInvalidTXT indicates that the UltraStar txt data could not be parsed.
	// It is usually accompanied with a line number that caused the error.
	TypeInvalidTXT = ProblemTypeDomain + "/invalid-ultrastar-txt"
)

// InvalidUltraStarTXT generates an error indicating that the UltraStar data in the request could not be parsed.
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
