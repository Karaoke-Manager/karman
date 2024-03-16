package apierror

import (
	"net/http"
)

const (
	// TypeInvalidJobState indicates that an action could not be performed because the referenced job
	// was in an invalid state.
	TypeInvalidJobState = ProblemTypeDomain + "invalid-job-state"
)

// InvalidJobState generates an error indicating that a job was in an invalid state.
// The error will have the provided detail message.
func InvalidJobState(detail string) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeInvalidJobState,
		Title:  "Invalid Job State",
		Status: http.StatusConflict,
		Detail: detail,
	}
}
