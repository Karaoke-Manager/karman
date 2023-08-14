package render

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// EncodeFunc is the signature of encoder functions that can be registered via [RegisterEncoder].
// An encoder function takes the following values:
// - w is the writer to which the response should be written.
// - v is the value that should be encoded.
//
// If an error occurs during encoding the encoder function may return it.
// Keep in mind however that at this point in the request lifecycle part of the response may already have been sent, so
// it is unlikely that the error can be handled in a meaningful manner.
//
// An encoder function may panic if the type of v is not compatible.
//
// Encoding functions must be safe for concurrent use.
type EncodeFunc func(w io.Writer, v any) error

// encoders contains all encoding functions registered via [RegisterEncoder].
var encoders = make(map[string]EncodeFunc)

// RegisterEncoder registers the given encoder function for the specified media types.
// Encoder functions can only be registered for types without parameters.
// Violations will cause a panic.
//
// It is not possible to register multiple encoders for the same type.
// A later registration will always take precedence.
//
// RegisterEncoder is not safe for concurrent use.
func RegisterEncoder(e EncodeFunc, mediaTypes ...string) {
	for _, t := range mediaTypes {
		mType := mediatype.MustParse(t)
		if len(mType.Parameters()) > 0 {
			panic("render: encoders can only be registered for types without parameters.")
		}
		encoders[mType.FullType()] = e
	}
}

// DefaultMediaType is the fallback media type that will be used by [Render] and [Respond] if no content type negotiation
// has been performed.
// It is usually recommended to use the content negotiation mechanisms to derive the response content type.
// This default mainly exists to lower the barrier of using this package,
// so that you can start using it without having to dive into the content type negotiation stuff.
var DefaultMediaType = mediatype.ApplicationJSON

// The Responder interface can be implemented by types that are intended to be passed to the [Render] and [Respond] functions.
// If a value implements this interface it gets a chance to do last-minute transformations or
// even provide a completely different replacement object.
//
// See the [Respond] function for details.
type Responder interface {
	// PrepareResponse is called before the receiver gets encoded as a response to a request.
	// This method may be implemented to do different things:
	//  - Set response headers commonly associated with this type of value.
	//  - Re-negotiate the response content type (this is especially useful for error responses).
	//  - Provide a completely different value that should be serialized.
	//
	// PrepareResponse is not invoked recursively.
	// The replacement value returned by this method is used as-is.
	// No further transformations will be made to it.
	PrepareResponse(w http.ResponseWriter, r *http.Request) any
}

// Respond writes v to w using a previously registered encoder.
//
// If v implements the [Responder] interface, the first thing this function does is call the [Responder.PrepareResponse] method,
// giving v the opportunity to do last-minute transformations.
// Only v itself is considered for this mechanism.
// If v is a struct value and any child values implement [Responder] it does not have any effect.
//
// Respond then determines the intended content type of the response.
// This is usually done via prior content type negotiation in the request chain.
// If you want to use a specific response format you can override the content type negotiation using [SetNegotiatedContentType].
// If no content negotiation has happened or the negotiation did not yield an acceptable media type,
// the response header Content-Type of w is inspected.
// If content type negotiation yielded a wildcard type, the response headers may also be used
// to determine a concrete Content-Type for the response.
// If this process yields a non-concrete media type, [Respond] panics.
// If no other means are available, the [DefaultMediaType] is used.
//
// Using the intended content type of the response Respond will then choose an encoder previously
// registered using [RegisterEncoder]. An encoder is chosen by the following rules:
//   - If an encoder has been registered for the exact type, this encoder is chosen.
//   - If an encoder has been registered for the subtype suffix, that encoder is chosen.
//   - If an encoder has been registered for a wildcard subtype of the major type, that encoder is chosen.
//   - If an encoder has been registered for "*/*", we chose this catchall encoder.
//
// In a last step before encoding v using the encoder Respond sets the Content-Type of the response (if unset) to the
// intended content type (or rather the concrete type chosen with the encoder).
// It a status code has been set via [SetStatus], that status code is sent before the encoding process is started.
func Respond(w http.ResponseWriter, r *http.Request, v any) error {
	if rsp, ok := v.(Responder); ok {
		v = rsp.PrepareResponse(w, r)
	}
	// Content negotiation usually happens in the middleware.
	// If a negotiated media type does not exist here, we use a default.
	contentType, _ := GetNegotiatedContentType(r)
	if !contentType.IsConcrete() {
		// If the result of content type negotiation is inconclusive,
		// we use the Content-Type response header to get more information.
		t, err := mediatype.Parse(w.Header().Get("Content-Type"))
		if err != nil && (contentType.IsNil() || contentType.Includes(t)) {
			contentType = t
		}
	}
	if !contentType.IsConcrete() {
		if contentType.IsNil() || contentType.Includes(DefaultMediaType) {
			// no content type negotiation has been done or the default type is more concrete
			contentType = DefaultMediaType
		}
	}
	if contentType.IsNil() {
		panic("render: no content type provided")
	}
	if !contentType.IsConcrete() {
		panic("render: content type must be concrete")
	}

	e := selectByMediaType(encoders, contentType)
	if e == nil {
		panic(fmt.Sprintf("no encoder available for %q", contentType.String()))
	}

	// Only set the content type if it hasn't been set
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", contentType.String())
	}
	// Set the status if available
	if status, ok := GetStatus(r); ok {
		w.WriteHeader(status)
	}
	return e(w, v)
}

// NoContent is a convenience function that writes HTTP 204 "No Content" to w.
// This function bypasses the entire rendering mechanism and mainly exists for consistency reasons, so you can write
// render.NoContent(w, r) as you would render.Render(w, r, v).
func NoContent(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
