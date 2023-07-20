package apierror

import (
	"encoding/json"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"mime"
	"net/http"
)

type ProblemDetails struct {
	Type     string `json:"type,omitempty" xml:"type,omitempty"`
	Title    string `json:"title,omitempty" xml:"title,omitempty"`
	Status   int    `json:"status,omitempty" xml:"status,omitempty"`
	Detail   string `json:"detail,omitempty" xml:"detail,omitempty"`
	Instance string `json:"instance,omitempty" xml:"instance,omitempty"`

	Fields  map[string]any `json:"-"`
	Headers http.Header    `json:"-"`
}

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

func (p *ProblemDetails) Error() string {
	return p.Title
}

func (p *ProblemDetails) Render(w http.ResponseWriter, r *http.Request) error {
	for key, values := range p.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if p.Type == "" || p.Type == "about:blank" || p.Title == "" {
		p.Title = http.StatusText(p.Status)
	}
	render.Status(r, p.Status)
	switch render.GetResponseFormat(r) {
	case render.FormatXML:
		render.ContentType(r, mime.FormatMediaType("application/problem+xml", map[string]string{"charset": "utf-8"}))
	default:
		render.ContentType(r, mime.FormatMediaType("application/problem+json", map[string]string{"charset": "utf-8"}))
	}
	return nil
}

func HttpStatus(status int) *ProblemDetails {
	return &ProblemDetails{Status: status}
}
