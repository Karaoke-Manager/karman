package schema

import (
	"github.com/Karaoke-Manager/karman/pkg/render"
)

type List[T render.Renderer] struct {
	render.NopRenderer
	Count  int `json:"count"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
	Items  []T `json:"items"`
}
