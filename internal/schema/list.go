package schema

import (
	"github.com/Karaoke-Manager/karman/pkg/render"
	"net/http"
)

type List[T render.Renderer] struct {
	Count  int   `json:"count"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
	Total  int64 `json:"total"`
	Items  []T   `json:"items"`
}

func (l *List[T]) Render(http.ResponseWriter, *http.Request) error {
	l.Count = len(l.Items)
	return nil
}
