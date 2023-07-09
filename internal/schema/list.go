package schema

import (
	"github.com/Karaoke-Manager/karman/pkg/render"
)

type List[T render.Renderer] struct {
	render.NopRenderer
	Items []T
}
