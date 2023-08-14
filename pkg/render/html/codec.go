// Package html implements a response encoder for HTML data.
// When this package is imported the encoder is automatically registered.
package html

import (
	"fmt"
	"io"
	"reflect"

	"github.com/Karaoke-Manager/server/pkg/render"
)

func init() {
	render.RegisterEncoder(Encode, "text/html", "application/xhtml+xml")
}

// Encode writes a string to the response.
func Encode(w io.Writer, v any) (err error) {
	switch h := v.(type) {
	case []byte:
		_, err = w.Write(h)
	case string:
		_, err = w.Write([]byte(h))
	case fmt.Stringer:
		_, err = w.Write([]byte(h.String()))
	case interface{ HTML() string }:
		_, err = w.Write([]byte(h.HTML()))
	case io.Reader:
		var b []byte
		b, err = io.ReadAll(h)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
	default:
		panic(fmt.Sprintf("cannot convert value of type %v to html", reflect.TypeOf(v).Kind()))
	}
	return
}
