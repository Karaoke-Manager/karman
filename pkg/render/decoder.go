package render

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"

	"github.com/ajg/form"
)

// Decode is a package-level variable set to our default Decoder. We do this
// because it allows you to set render.Decode to another function with the
// same function signature, while also utilizing the render.Decoder() function
// itself. Effectively, allowing you to easily add your own logic to the package
// defaults. For example, maybe you want to impose a limit on the number of
// bytes allowed to be read from the request body.
var Decode = DefaultDecoder

// DefaultDecoder detects the correct decoder for use on an HTTP request and
// marshals into a given interface.
//
// If the request does not contain a body, no decoding takes place.
func DefaultDecoder(r *http.Request, v any) (err error) {
	// Do nothing if request body is empty
	rd := bufio.NewReader(r.Body)
	_, err = rd.Peek(1)
	if errors.Is(err, io.EOF) {
		return nil
	}
	switch GetRequestFormat(r) {
	case FormatJSON:
		err = DecodeJSON(rd, v)
	case FormatXML:
		err = DecodeXML(rd, v)
	case FormatForm:
		err = DecodeForm(rd, v)
	case FormatEmpty:
		// No content type specified, guess JSON. Use a middleware if the
		// presence of a content type is required.
		err = DecodeJSON(rd, v)
	default:
		err = ErrUnsupportedFormat
	}

	return
}

// DecodeJSON decodes a given reader into an interface using the json decoder.
func DecodeJSON(r io.Reader, v any) error {
	defer io.Copy(io.Discard, r)
	return json.NewDecoder(r).Decode(v)
}

// DecodeXML decodes a given reader into an interface using the xml decoder.
func DecodeXML(r io.Reader, v any) error {
	defer io.Copy(io.Discard, r)
	return xml.NewDecoder(r).Decode(v)
}

// DecodeForm decodes a given reader into an interface using the form decoder.
func DecodeForm(r io.Reader, v any) error {
	decoder := form.NewDecoder(r)
	return decoder.Decode(v)
}
