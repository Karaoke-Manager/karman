// Package form implements a request decoder for form data.
// When this package is imported the decoder is automatically registered.
package form

import (
	"io"

	"github.com/Karaoke-Manager/server/pkg/mediatype"
	"github.com/Karaoke-Manager/server/pkg/render"
	"github.com/ajg/form"
)

func init() {
	render.RegisterDecoder(Decode, "application/x-www-form-urlencoded")
}

// Decode reads form-encoded data from r.
func Decode(r io.Reader, _ mediatype.MediaType, v any) (err error) {
	defer func() {
		_, cErr := io.Copy(io.Discard, r)
		if err == nil {
			err = cErr
		}
	}()
	decoder := form.NewDecoder(r)
	return decoder.Decode(v)
}
