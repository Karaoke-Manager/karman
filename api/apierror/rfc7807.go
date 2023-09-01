package apierror

import (
	"encoding/json"
	"net/http"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// ProblemDetails implements [RFC 7807] "Problem Details for HTTP APIs".
// A ProblemDetails value is intended to be used for error reporting from the Karman API.
// Problem details are usually categorized by their type which is a URI.
// This package defines utility functions and constants for typical errors.
// All error reporting of the Karman API should be done using values of this type.
//
// A ProblemDetails value is intended to be serialized using the
// [github.com/Karaoke-Manager/karman/pkg/render] package.
//
// [RFC 7807]: https://datatracker.ietf.org/doc/html/rfc7807
type ProblemDetails struct {
	// A URI reference that identifies the problem type.
	// When dereferenced it should provide human-readable documentation for the problem type.
	// When empty its value is assumed to be "about:blank".
	Type string `json:"type,omitempty" xml:"type,omitempty"`
	// A short, human-readable summary of the problem type.
	// It should not change from occurrence to occurrence of the problem, except for purposes of localization
	Title string `json:"title,omitempty" xml:"title,omitempty"`
	// The HTTP status code generated by the origin server for this occurrence of the problem.
	Status int `json:"status,omitempty" xml:"status,omitempty"`
	// A human-readable explanation specific to this occurrence of the problem.
	Detail string `json:"detail,omitempty" xml:"detail,omitempty"`
	// A URI reference that identifies the specific occurrence of the problem.
	// It may or may not yield further information if dereferenced.
	Instance string `json:"instance,omitempty" xml:"instance,omitempty"`

	// Additional fields providing details about the cause of the problem.
	// When rendering to JSON these will be included at the top level.
	Fields map[string]any `json:"-"`
	// Additional HTTP headers that should be set when using the render package.
	Headers http.Header `json:"-"`
}

// MarshalJSON encodes p into a JSON structure.
// Rendering to JSON will include ProblemDetails.Fields at the top level.
// If a field has the same key as one of the properties, the property will take precedence.
func (p *ProblemDetails) MarshalJSON() ([]byte, error) {
	data := make(map[string]any, 5+len(p.Fields))
	for key, value := range p.Fields {
		data[key] = value
	}
	if p.Type != "" {
		data["type"] = p.Type
	}
	if p.Title != "" {
		data["title"] = p.Title
	}
	if p.Status != 0 {
		data["status"] = p.Status
	}
	if p.Detail != "" {
		data["detail"] = p.Detail
	}
	if p.Instance != "" {
		data["instance"] = p.Instance
	}
	return json.Marshal(data)
}

// IsDefaultType indicates whether p.Type is the default problem type for plain status codes.
// If this method returns true p.Type should be considered equal to "about:blank".
func (p *ProblemDetails) IsDefaultType() bool {
	return p.Type == "" || p.Type == "about:blank"
}

// UnmarshalJSON decodes data into p.
// This counterpart to ProblemDetails.MarshalJSON puts all unknown fields into p.Fields.
func (p *ProblemDetails) UnmarshalJSON(data []byte) error {
	type problemDetails ProblemDetails
	if err := json.Unmarshal(data, (*problemDetails)(p)); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &p.Fields); err != nil {
		return err
	}
	delete(p.Fields, "type")
	delete(p.Fields, "title")
	delete(p.Fields, "status")
	delete(p.Fields, "detail")
	delete(p.Fields, "instance")
	return nil
}

// Error implements the error interface for ProblemDetails.
func (p *ProblemDetails) Error() string {
	return p.Title
}

// Render prepares p to be written to w.
// This method prepares p with some default values.
func (p *ProblemDetails) Render(_ http.ResponseWriter, _ *http.Request) error {
	if p.Type == "" || p.Type == "about:blank" || p.Title == "" {
		p.Title = http.StatusText(p.Status)
	}

	return nil
}

// PrepareResponse implements the [render.Responder] interface.
// This implementation writes the headers of p, sets a status code and negotiates an appropriate content type.
func (p *ProblemDetails) PrepareResponse(w http.ResponseWriter, r *http.Request) any {
	for key, values := range p.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	render.SetStatus(r, p.Status)
	render.MustNegotiateContentType(r, mediatype.ApplicationProblemJSON, mediatype.ApplicationProblemXML, mediatype.ApplicationJSON, mediatype.ApplicationXML)
	return p
}

// HTTPStatus returns a ProblemDetails value representing the specified status.
// When rendering the value will have its title set to the default status text for status.
func HTTPStatus(status int) *ProblemDetails {
	return &ProblemDetails{Status: status}
}
