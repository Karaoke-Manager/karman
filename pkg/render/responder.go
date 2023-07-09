package render

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

var (
	ErrRequestTimeout = errors.New("request context timed out")
)

// M is a convenience alias for quickly building a map structure that is going
// out to a responder. Just a short-hand.
// type M map[string]any

// Respond is a package-level variable set to our default Responder. We do this
// because it allows you to set render.Respond to another function with the
// same function signature, while also utilizing the render.Responder() function
// itself. Effectively, allowing you to easily add your own logic to the package
// defaults. For example, maybe you want to test if v is an error and respond
// differently, or log something before you respond.
var Respond = DefaultResponder

// DefaultResponder handles streaming JSON and XML responses, automatically
// setting the Content-Type based on request headers. It will default to a JSON
// response.
func DefaultResponder(w http.ResponseWriter, r *http.Request, v any) error {
	format := GetResponseFormat(r)
	if v != nil {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Chan:
			switch format {
			case FormatEventStream:
				return channelEventStream(w, r, v)
			default:
				v = channelIntoSlice(w, r, v)
			}
		}
	}

	// Format response based on request Accept header (or context)
	switch format {
	case FormatPlainText:
		return PlainText(w, r, v)
	case FormatHTML:
		return HTML(w, r, v)
	case FormatJSON:
		return JSON(w, r, v)
	case FormatXML:
		return XML(w, r, v)
	case FormatData:
		return Data(w, r, v)
	default:
		return JSON(w, r, v)
	}
}

// NoContent returns a HTTP 204 "No Content" response.
func NoContent(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// PlainText writes a string to the response.
func PlainText(w http.ResponseWriter, r *http.Request, v any) (err error) {
	writeHeader(w, r, FormatPlainText)
	// TODO: Maybe find other ways of converting v to bytes
	switch s := v.(type) {
	case string:
		_, err = w.Write([]byte(s))
	case []byte:
		_, err = w.Write(s)
	case interface{ String() string }:
		_, err = w.Write([]byte(s.String()))
	case io.Reader:
		var b []byte
		b, err = io.ReadAll(s)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
	default:
		panic(fmt.Sprintf("cannot convert value of type %v to string", reflect.TypeOf(v).Kind()))
	}
	return
}

// Data writes raw bytes to the response.
func Data(w http.ResponseWriter, r *http.Request, v any) (err error) {
	writeHeader(w, r, FormatData)
	// TODO: Maybe find other ways of converting v to bytes
	switch b := v.(type) {
	case []byte:
		_, err = w.Write(b)
	case string:
		_, err = w.Write([]byte(b))
	case interface{ String() string }:
		_, err = w.Write([]byte(b.String()))
	case io.Reader:
		var bs []byte
		bs, err = io.ReadAll(b)
		if err != nil {
			return err
		}
		_, err = w.Write(bs)
	default:
		panic(fmt.Sprintf("cannot convert value of type %v to []byte", reflect.TypeOf(v).Kind()))
	}
	return
}

// HTML writes a string to the response, setting the Content-Type as text/html.
func HTML(w http.ResponseWriter, r *http.Request, v any) (err error) {
	writeHeader(w, r, FormatHTML)
	switch h := v.(type) {
	case []byte:
		_, err = w.Write(h)
	case string:
		_, err = w.Write([]byte(h))
	case interface{ String() string }:
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

// JSON marshals 'v' to JSON, automatically escaping HTML.
func JSON(w http.ResponseWriter, r *http.Request, v any) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		panic(fmt.Errorf("cannot encode JSON: %e", err))
	}

	writeHeader(w, r, FormatJSON)
	_, err := w.Write(buf.Bytes())
	return err
}

// XML marshals 'v' to XML. It will automatically prepend a generic XML header
// (see encoding/xml.Header) if one is not found in the first 100 bytes of 'v'.
func XML(w http.ResponseWriter, r *http.Request, v any) (err error) {
	b, err := xml.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("cannot encode XML: %e", err))
	}

	writeHeader(w, r, FormatXML)

	// Try to find <?xml header in first 100 bytes (just in case there're some XML comments).
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

// writeHeader writes the response header based on the request context and the
// specified format. If a ContextKeyContentType is present in the request the
// value is used, otherwise the preferred content type of format is used.
// If ContextKeyStatus is present in the request, it is used as the response
// code.
func writeHeader(w http.ResponseWriter, r *http.Request, format Format) {
	if contentType, ok := r.Context().Value(ContextKeyContentType).(string); ok {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", format.PreferredContentType())
	}
	if status, ok := r.Context().Value(ContextKeyStatus).(int); ok {
		w.WriteHeader(status)
	}
}

// channelEventStream streams the data from v as response. v must be a channel
// sending json-encodeable values.
func channelEventStream(w http.ResponseWriter, r *http.Request, v any) error {
	// FIXME: There are some magic strings here that we might want to investigate.
	if reflect.TypeOf(v).Kind() != reflect.Chan {
		panic(fmt.Sprintf("render: event stream expects a channel, not %v", reflect.TypeOf(v).Kind()))
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	if r.ProtoMajor == 1 {
		// An endpoint MUST NOT generate an HTTP/2 message containing connection-specific header fields.
		// Source: RFC7540
		w.Header().Set("Connection", "keep-alive")
	}

	w.WriteHeader(http.StatusOK)

	ctx := r.Context()
	for {
		switch chosen, recv, ok := reflect.Select([]reflect.SelectCase{
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(v)},
		}); chosen {
		case 0: // equivalent to: case <-ctx.Done()
			_, _ = w.Write([]byte("event: error\ndata: {\"error\":\"Server Timeout\"}\n\n"))
			return ErrRequestTimeout

		default: // equivalent to: case v, ok := <-stream
			if !ok {
				_, err := w.Write([]byte("event: EOF\n\n"))
				return err
			}
			v := recv.Interface()

			// Build each channel item.
			if rv, ok := v.(Renderer); ok {
				err := renderer(w, r, rv)
				if err != nil {
					v = err
				} else {
					v = rv
				}
			}

			b, err := json.Marshal(v)
			if err != nil {
				_, _ = w.Write([]byte(fmt.Sprintf("event: error\ndata: {\"error\":\"%v\"}\n\n", err)))
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				continue
			}
			_, err = w.Write([]byte(fmt.Sprintf("event: data\ndata: %s\n\n", b)))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			if err != nil {
				return err
			}
		}
	}
}

// channelIntoSlice buffers channel data into a slice.
func channelIntoSlice(w http.ResponseWriter, r *http.Request, from interface{}) interface{} {
	ctx := r.Context()

	var to []interface{}
	for {
		switch chosen, recv, ok := reflect.Select([]reflect.SelectCase{
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(from)},
		}); chosen {
		case 0: // equivalent to: case <-ctx.Done()
			http.Error(w, "Server Timeout", http.StatusGatewayTimeout)
			return nil

		default: // equivalent to: case v, ok := <-stream
			if !ok {
				return to
			}
			v := recv.Interface()

			// Render each channel item.
			if rv, ok := v.(Renderer); ok {
				err := renderer(w, r, rv)
				if err != nil {
					v = err
				} else {
					v = rv
				}
			}

			to = append(to, v)
		}
	}
}
