package schema

import (
	"net/http"
	"strconv"

	"github.com/Karaoke-Manager/karman/pkg/render"
)

// List is a generic schema type for list responses.
// Apart from the actual slice of items a List contains information about the size of the underlying query result.
type List[T render.Renderer] struct {
	// The number of elements in this list response.
	// Equal to len(l.Items), less than or equal to Total.
	Count int

	// The Offset of this list within the underlying collection.
	Offset int64

	// The Limit from the request. Greater or equal to Count.
	Limit int

	// The Total number of elements in the underlying collection.
	Total int64

	// A slice of items in this list.
	Items []T
}

// Render makes sure that l.Count is correctly set.
// Render implements the render.Renderer interface.
func (l *List[T]) Render(http.ResponseWriter, *http.Request) error {
	l.Count = len(l.Items)
	return nil
}

// PrepareResponse generates the actual response list from l.
// This method also sets pagination headers.
func (l *List[T]) PrepareResponse(w http.ResponseWriter, _ *http.Request) any {
	w.Header().Set("Pagination-Count", strconv.Itoa(l.Count))
	w.Header().Set("Pagination-Offset", strconv.FormatInt(l.Offset, 10))
	w.Header().Set("Pagination-Limit", strconv.Itoa(l.Limit))
	w.Header().Set("Pagination-Total", strconv.FormatInt(l.Total, 10))
	return l.Items
}
