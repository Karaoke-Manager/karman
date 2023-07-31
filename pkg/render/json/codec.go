// Package json implements a request decoder and a response encoder for JSON data.
// When this package is imported the encoder and decoder are automatically registered.
package json

import (
	"encoding/json"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"io"
)

func init() {
	render.RegisterDecoder(Decode, "application/json", "application/*+json")
	render.RegisterEncoder(Encode, "application/json", "application/*+json")
}

// Decode reads JSON data from r into v.
func Decode(r io.Reader, _ mediatype.MediaType, v any) (err error) {
	defer func() {
		_, cErr := io.Copy(io.Discard, r)
		if err == nil {
			err = cErr
		}
	}()
	return json.NewDecoder(r).Decode(v)
}

// Encode marshals 'v' to JSON, automatically escaping HTML.
func Encode(w io.Writer, v any) error {
	e := json.NewEncoder(w)
	e.SetEscapeHTML(true)
	return e.Encode(v)
}
