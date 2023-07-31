package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"io"
)

func init() {
	render.RegisterDecoder(Decode, "application/xml", "text/xml", "text/html", "application/*+xml")
	render.RegisterEncoder(Encode, "application/xml", "text/xml", "application/*+xml")
}

// Decode reads data from r into v.
func Decode(r io.Reader, _ mediatype.MediaType, v any) (err error) {
	defer func() {
		_, cErr := io.Copy(io.Discard, r)
		if err == nil {
			err = cErr
		}
	}()
	return xml.NewDecoder(r).Decode(v)
}

// Encode marshals 'v' to Encode.
// It will automatically prepend a generic Encode header (see encoding/xml.Header) if one is not found in the first 100 bytes of 'v'.
func Encode(w io.Writer, v any) (err error) {
	b, err := xml.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("cannot encode XML: %e", err))
	}

	// Try to find <?xml header in first 100 bytes (just in case there are some XML comments).
	findHeaderUntil := len(b)
	if findHeaderUntil > 100 {
		findHeaderUntil = 100
	}
	if !bytes.Contains(b[:findHeaderUntil], []byte("<?xml")) {
		// No header found. Print it out first.
		_, err = w.Write([]byte(xml.Header))
		if err != nil {
			return
		}
	}

	_, err = w.Write(b)
	return
}
