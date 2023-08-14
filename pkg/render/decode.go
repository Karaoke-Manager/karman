package render

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Karaoke-Manager/server/pkg/mediatype"
)

// DecodeFunc is the signature of a decoding function that can be registered via [RegisterDecoder].
// A decoding function takes the following values:
//   - r is an [io.Reader] for the request body. This is the data to be decoded.
//   - mediaType is the Content-Type of the request that caused this decoder to be chosen.
//     If a decoder is registered for multiple media types this allows you to disambiguate based on the concrete type.
//     Note that mediaType is basically just the Content-Type of the request, which may be a wildcard type.
//   - v is the receiving value into which the data from r should be decoded.
//
// If an error occurs during decoding, an error should be returned.
// A decoding function may panic if the type of v is not appropriate (e.g. not a pointer type).
// A decoding function should not panic if it cannot handle the mediaType or data.
//
// Decoding functions must be safe for concurrent use.
type DecodeFunc func(r io.Reader, mediaType mediatype.MediaType, v any) error

// decoders stores all registered decoder functions.
var decoders = make(map[string]DecodeFunc)

// RegisterDecoder registers the decoding function d for the specified media types.
// The provided types can be concrete types or wildcard types that indicate that a decoder may be used for
// any types matching that wildcard.
// If there are decoders registered for concrete types as well as wildcards, the more concrete types always take precedence.
//
// It is not possible to register multiple decoders for the same type.
// If you do register multiple decoders the later registration will take precedence.
// The mediaTypes cannot have parameters.
//
// RegisterDecoder is not safe for concurrent use.
func RegisterDecoder(d DecodeFunc, mediaTypes ...string) {
	for _, rawType := range mediaTypes {
		mType, err := mediatype.Parse(rawType)
		if err != nil {
			panic(fmt.Sprintf("render: invalid media type: %q", rawType))
		}
		if len(mType.Parameters()) > 0 {
			panic("render: decoders can only be registered for media types without parameters")
		}
		decoders[mType.FullType()] = d
	}
}

// These are known errors that happen during decoding of requests.
var (
	// ErrMissingContentType indicates that the request did not specify a Content-Type header.
	// If you want to default to a fixed decoder (e.g. JSON as a default) use the [DefaultRequestContentType] middleware.
	ErrMissingContentType = errors.New("missing content-type header")
	// ErrInvalidContentType indicates that the Content-Type header was present but not correctly formatted.
	ErrInvalidContentType = errors.New("invalid content-type header")
	// ErrNoMatchingDecoder indicates that no decoder has been registered for the Content-Type provided by the request.
	ErrNoMatchingDecoder = errors.New("no matching decoder has been registered")
)

// Decode reads the request body, decodes it and stores it into v.
//
// The heavy lifting of the Decode function is done by a decoder, previously registered via [RegisterDecoder].
// The Decode function inspects the Content-Type of r and then chooses an appropriate decoder to decode r.
// A decoder is chosen by the following priorities:
//   - If a decoder has been registered for the exact Content-Type of r, this decoder is chosen.
//   - If a decoder has been registered for the subtype suffix of the request's Content-Type, that decoder is chosen.
//   - If a decoder has been registered for a wildcard subtype of the major type of r's Content-Type, that decoder is chosen.
//   - If a decoder has been registered for "*/*", this catchall decoder is used.
//
// If no decoder is found, ErrNoMatchingDecoder is returned without reading the request body.
//
// In your handling of a request you should make sure that the value you provide is compatible with the possible decoders.
// It is a programmer error if the type of v is incompatible with the decoder, causing Decode to panic.
func Decode(r *http.Request, v any) error {
	h := r.Header.Get("Content-Type")
	if strings.TrimSpace(h) == "" {
		return fmt.Errorf("render: cannot decode request: %w", ErrMissingContentType)
	}
	mediaType, _ := mediatype.Parse(h)
	if mediaType.IsNil() {
		return fmt.Errorf("render: cannot parse %q: %w", r.Header.Get("Content-Type"), ErrInvalidContentType)
	}
	d := selectByMediaType(decoders, mediaType)
	if d == nil {
		return fmt.Errorf("render: no decorder for content type %q: %w", mediaType.FullType(), ErrNoMatchingDecoder)
	}
	return d(r.Body, mediaType, v)
}

// selectByMediaType is a little helper function that implements selection of a registered encoder or decoder based on a
// concrete media type.
// This function understands wildcard registrations.
func selectByMediaType[T any](m map[string]T, t mediatype.MediaType) T {
	v, ok := m[t.FullType()]
	if !ok {
		suffix := t.SubtypeSuffix()
		if suffix == "" {
			suffix = t.Subtype()
		}
		v, ok = m[t.Type()+"/*+"+suffix]
	}
	if !ok {
		v, ok = m[t.Type()+"/*"]
	}
	if !ok {
		v = m["*/*"]
	}
	return v
}

// DefaultRequestContentType returns a middleware that provides requests with a default Content-Type.
// Requests that do not have a Content-Type header set, will have it set to v.
//
// Use this middleware if you want to be able to [Decode] requests that do not specify a Content-Type.
func DefaultRequestContentType(v string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") == "" {
				r.Header.Set("Content-Type", v)
			}
		}
		return http.HandlerFunc(fn)
	}
}

// SetRequestContentType returns a middleware that overwrites the Content-Type header of all requests with v.
//
// Use this middleware if you want to disregard the provided Content-Type header and use the same decoder for all requests
// (e.g. if you want to only support JSON).
// Usually a better approach is to use a middleware that responds with a 415 status code to unsupported Content-Type values.
func SetRequestContentType(v string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("Content-Type", v)
		}
		return http.HandlerFunc(fn)
	}
}
