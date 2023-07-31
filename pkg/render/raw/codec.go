package raw

import (
	"fmt"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"io"
)

func init() {
	render.RegisterEncoder(Encode, "application/octet-stream", "text/plain", "*/*")
	render.RegisterDecoder(Decode, "application/octet-stream", "text/plain", "*/*")
}

// Encode writes raw bytes to the response.
func Encode(w io.Writer, v any) (err error) {
	switch b := v.(type) {
	case []byte:
		_, err = w.Write(b)
	case string:
		_, err = w.Write([]byte(b))
	case fmt.Stringer:
		_, err = w.Write([]byte(b.String()))
	case io.Reader:
		var bs []byte
		bs, err = io.ReadAll(b)
		if err != nil {
			return err
		}
		_, err = w.Write(bs)
	default:
		panic(fmt.Sprintf("cannot encode value of type %T", v))
	}
	return
}

// Decode reads binary data into v.
func Decode(r io.Reader, _ mediatype.MediaType, v any) error {
	switch x := v.(type) {
	case io.Writer:
		_, err := io.Copy(x, r)
		return err
	case []byte, *[]byte, *string:
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	switch x := v.(type) {
	case []byte:
		copy(x, data)
	case *[]byte:
		*x = data
	case *string:
		*x = string(data)
	}
	return err
}
