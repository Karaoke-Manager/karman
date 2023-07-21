package schema

import (
	"github.com/Karaoke-Manager/karman/pkg/render"
	"net/http"
)

// List is a generic schema type for list responses.
// Apart from the actual slice of items a List contains information about the size of the underlying query result.
type List[T render.Renderer] struct {
	// The number of elements in this list response.
	// Equal to len(l.Items), less than or equal to Total.
	Count int `json:"count"`

	// The Offset of this list within the underlying collection.
	Offset int `json:"offset"`

	// The Limit from the request. Greater or equal to Count.
	Limit int `json:"limit"`

	// The Total number of elements in the underlying collection.
	Total int64 `json:"total"`

	// A slice of items in this list.
	Items []T `json:"items"`
}

// Render makes sure that l.Count is correctly set.
// Render implements the render.Renderer interface.
func (l *List[T]) Render(http.ResponseWriter, *http.Request) error {
	l.Count = len(l.Items)
	return nil
}
