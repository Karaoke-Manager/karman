package render

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported format")
)

// Format is an enumeration of formats supported by the render package.
type Format int

// ContentTypes handled by this package.
const (
	FormatUnsupported = iota
	FormatEmpty
	FormatPlainText
	FormatHTML
	FormatJSON
	FormatXML
	FormatForm
	FormatData
	FormatEventStream
)

// GetFormatFromContentType returns a Format value based on the specified
// content type or FormatUnsupported if the content type is not supported.
func GetFormatFromContentType(s string) Format {
	mediaType, _, err := mime.ParseMediaType(s)
	switch mediaType {
	case "":
		if err != nil {
			return FormatUnsupported
		}
		return FormatEmpty
	case "text/plain":
		return FormatPlainText
	case "text/html", "application/xhtml+xml":
		return FormatHTML
	case "application/json", "text/javascript":
		return FormatJSON
	case "text/xml", "application/xml":
		return FormatXML
	case "application/x-www-form-urlencoded":
		return FormatForm
	case "application/octet-stream":
		return FormatData
	case "text/event-stream":
		return FormatEventStream
	default:
		return FormatUnsupported
	}
}

// PreferredContentType returns the preferred content type of format f.
func (f Format) PreferredContentType() string {
	switch f {
	case FormatUnsupported, FormatEmpty:
		return ""
	case FormatPlainText:
		return mime.FormatMediaType("text/plain", map[string]string{"charset": "utf-8"})
	case FormatHTML:
		return mime.FormatMediaType("text/html", map[string]string{"charset": "utf-8"})
	case FormatJSON:
		return mime.FormatMediaType("application/json", map[string]string{"charset": "utf-8"})
	case FormatXML:
		return mime.FormatMediaType("application/xml", map[string]string{"charset": "utf-8"})
	case FormatForm:
		return mime.FormatMediaType("application/x-www-form-urlencoded", nil)
	case FormatData:
		return mime.FormatMediaType("application/octet-stream", nil)
	case FormatEventStream:
		return mime.FormatMediaType("application/event-stream", nil)
	default:
		return ""
	}
}

// SetRequestFormat is a middleware that forces the render package to decode
// request bodies in the specified format. By default, requests are decoded
// based on the request content type.
func SetRequestFormat(format Format) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), ContextKeyRequestFormat, format))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// SetResponseFormat is a middleware that forces the render package to encode
// responses in the specified format. By default, responses are encoded based on
// the Accept header.
func SetResponseFormat(format Format) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), ContextKeyResponseFormat, format))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// SetFormat is a middleware that forces the render package to encode and decpde
// responses in the specified format. By default, responses are encoded based on
// the request headers.
func SetFormat(format Format) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ContextKeyRequestFormat, format)
			ctx = context.WithValue(ctx, ContextKeyResponseFormat, format)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// GetRequestFormat is a helper function that returns the Format for request
// data based on context or request headers.
func GetRequestFormat(r *http.Request) Format {
	if format, ok := r.Context().Value(ContextKeyRequestFormat).(Format); ok {
		return format
	}
	return GetFormatFromContentType(r.Header.Get("Content-Type"))
}

// GetResponseFormat is a helper function that returns the Format for response
// data based on context or request data.
func GetResponseFormat(r *http.Request) (format Format) {
	if format, ok := r.Context().Value(ContextKeyResponseFormat).(Format); ok {
		return format
	}

	// Parse request Accept header.
	header := r.Header.Get("Accept")
	if header == "" {
		return FormatEmpty
	}
	fields := strings.Split(header, ",")
	for _, field := range fields {
		format = GetFormatFromContentType(strings.TrimSpace(field))
		if format != FormatUnsupported && format != FormatEmpty {
			return format
		}
	}
	return FormatUnsupported
}
