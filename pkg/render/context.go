package render

// TODO: Maybe publish this as a standalone fork of the CHI repository.

import (
	"context"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"net/http"
)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer, so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey int

const (
	// contextKeyAcceptedMediaTypes is a context key that holds a parsed version of the Accept header.
	contextKeyAcceptedMediaTypes contextKey = iota

	// contextKeyNegotiatedMediaType is a context key that holds the preferred response content type
	// as determined by content type negotiation.
	contextKeyNegotiatedMediaType

	// contextKeyNotAcceptableHandler is a context key that holds an HTTP handler that is used to handle
	// failed content type negotiation.
	contextKeyNotAcceptableHandler

	// contextKeyStatus is a context key that holds the future HTTP response status code.
	contextKeyStatus
)

// SetNegotiatedContentType sets t as the negotiated content type in the request context.
// This overrides any previously stored negotiation result.
// There are two primary use cases for this function:
//  1. You have implemented your own content type negotiation and want to provide the resulting media type to the render package.
//  2. You want to override the result of the content type negotiation with another value.
//     This might be useful if you want to send error responses that are not compatible with any of the accepted media types.
//
// The stored content type can be extracted from the returned context via [GetNegotiatedContentType] and [MustGetNegotiatedContentType].
func SetNegotiatedContentType(r *http.Request, t mediatype.MediaType) {
	p, ok := r.Context().Value(contextKeyNegotiatedMediaType).(*mediatype.MediaType)
	if ok {
		// We expect this function to be used multiple times during the request lifecycle.
		// We avoid unnecessarily stacking contexts.
		*p = t
	} else {
		*r = *r.WithContext(context.WithValue(r.Context(), contextKeyNegotiatedMediaType, &t))
	}
}

// GetNegotiatedContentType fetches the result of content type negotiation from the context.
// If no content type negotiation has been performed yet, ok will be false.
// If content type negotiation has been performed but has not yielded any usable media types, t will be [mediatype.Nil].
func GetNegotiatedContentType(r *http.Request) (t mediatype.MediaType, ok bool) {
	var p *mediatype.MediaType
	p, ok = r.Context().Value(contextKeyNegotiatedMediaType).(*mediatype.MediaType)
	if p != nil {
		t = *p
	}
	return
}

// MustGetNegotiatedContentType works just like [GetNegotiatedContentType] but panics
// if no content type negotiation has been performed yet.
// Use this function if you can be sure that negotiation has been performed (e.g. by a middleware).
func MustGetNegotiatedContentType(r *http.Request) mediatype.MediaType {
	return r.Context().Value(contextKeyNegotiatedMediaType).(mediatype.MediaType)
}

// setNotAcceptableHandler sets handler in the context to be used by a [ContentTypeNegotiation] middleware.
func setNotAcceptableHandler(r *http.Request, handler http.Handler) {
	*r = *r.WithContext(context.WithValue(r.Context(), contextKeyNotAcceptableHandler, handler))
}

// getNotAcceptableHandler fetches a handler, previously set via setNotAcceptableHandler.
func getNotAcceptableHandler(r *http.Request) (h http.Handler, ok bool) {
	h, ok = r.Context().Value(contextKeyNotAcceptableHandler).(http.Handler)
	return
}

// SetStatus sets an HTTP response status code hint into request context.
// This context value is respected by the [Respond] function.
func SetStatus(r *http.Request, status int) {
	*r = *r.WithContext(context.WithValue(r.Context(), contextKeyStatus, status))
}

// GetStatus returns the HTTP status code previously set via [SetStatus].
// If no such code exists, ok will be false.
func GetStatus(r *http.Request) (v int, ok bool) {
	v, ok = r.Context().Value(contextKeyStatus).(int)
	return
}

// MustGetStatus works just like [GetStatus] but panics if no status value has been set.
func MustGetStatus(r *http.Request) int {
	return r.Context().Value(contextKeyStatus).(int)
}
