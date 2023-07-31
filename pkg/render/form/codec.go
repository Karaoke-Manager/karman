package form

import (
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/ajg/form"
	"io"
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
