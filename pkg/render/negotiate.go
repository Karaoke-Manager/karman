package render

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/Karaoke-Manager/server/pkg/mediatype"
)

// NotAcceptableHandler returns a middleware that registers h as a handler for failed content type negotiation.
// This is used by the [ContentTypeNegotiation] middleware.
func NotAcceptableHandler(h http.HandlerFunc) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			setNotAcceptableHandler(r, h)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// NotAcceptable invokes a handler, previously registered via the [NotAcceptableHandler] middleware.
// If no handler has been registered, this function panics.
// Use this function if you are (re-)negotiating content types.
func NotAcceptable(w http.ResponseWriter, r *http.Request) {
	// We could not determine a common media type
	h, _ := getNotAcceptableHandler(r)
	h.ServeHTTP(w, r)
}

// ContentTypeNegotiation returns a middleware that performs content type negotiation on each request
// using the provided available media types.
// If there is an intersection of accepted types provided in the request's Accept header and the types provided as available,
// this middleware will find it and use SetNegotiatedContentType to set the resulting type in the request context.
// If the intersection of types is empty (aka the content type negotiation failed) and the [NotAcceptableHandler] middleware
// has been used to set a handler for this case, the request is routed to that handler.
// If no such handler is defined, the request chain will continue (with a negotiated type of [mediatype.Nil]).
//
// The [Respond] and [Render] functions require content type negotiation to produce a concrete type.
// Be careful with using wildcard types as an available type as
// it may cause a panic if the wildcard is not resolved before the response is sent.
//
// This middleware will also add a Vary:Accept response header.
func ContentTypeNegotiation(available ...string) func(next http.Handler) http.Handler {
	availableTypes := make([]mediatype.MediaType, len(available))
	for i, v := range available {
		availableTypes[i] = mediatype.MustParse(v)
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			found := false
			for _, raw := range w.Header().Values("Vary") {
				for _, v := range strings.Split(raw, ",") {
					if strings.TrimSpace(strings.ToLower(v)) == "accept" {
						found = true
						break
					}
				}
			}
			if !found {
				w.Header().Add("Vary", "Accept")
			}

			t := NegotiateContentType(r, availableTypes...)
			if t.IsNil() {
				// We could not determine a common media type
				h, _ := getNotAcceptableHandler(r)
				if h != nil {
					h.ServeHTTP(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// GetAcceptedMediaTypes parses the Accept header of the request and returns the resulting list of media types.
// The parsed result is stored in the request context so that multiple invocations of this function are not terribly inefficient.
// The resulting list is sorted by media type priority.
func GetAcceptedMediaTypes(r *http.Request) mediatype.MediaTypes {
	l, ok := r.Context().Value(contextKeyAcceptedMediaTypes).(mediatype.MediaTypes)
	if !ok {
		l = mediatype.ParseList(r.Header.Get("Accept"))
		sort.Stable(l)
		*r = *r.WithContext(context.WithValue(r.Context(), contextKeyAcceptedMediaTypes, l))
	}
	return l
}

// NegotiateContentType performs content type negotiation using the Accept header of r and the available media types.
// The result of this operation will be the highest priority intersection of the accepted media types and the available media types.
// If no such intersection exists, the result will be [mediatype.Nil].
// In any case the result will be stored in the request context as the negotiated media type.
//
// The [Respond] and [Render] functions require content type negotiation to produce a concrete type.
// Be careful with using wildcard types as an available type as
// it may cause a panic if the wildcard is not resolved before the response is sent.
func NegotiateContentType(r *http.Request, available ...mediatype.MediaType) mediatype.MediaType {
	accepted := GetAcceptedMediaTypes(r)
	best := accepted.BestMatch(available...)
	SetNegotiatedContentType(r, best)
	return best
}

// MustNegotiateContentType works like [NegotiateContentType] but if the intersection of accepted and available media types is empty
// instead of [mediatype.Nil] the highest priority available type is chosen.
//
// You might want to use this function over [NegotiateContentType] if having no resulting content type is not an option
// (such as for error responses).
// This is usually not the primary content type negotiation but used as a re-negotiation if the primary negotiation was inconclusive.
//
// The [Respond] and [Render] functions require content type negotiation to produce a concrete type.
// Be careful with using wildcard types as an available type as
// it may cause a panic if the wildcard is not resolved before the response is sent.
func MustNegotiateContentType(r *http.Request, available ...mediatype.MediaType) mediatype.MediaType {
	accepted := GetAcceptedMediaTypes(r)
	best := accepted.BestMatch(available...)
	if best.IsNil() {
		best = mediatype.MediaTypes(available).Best()
	}
	SetNegotiatedContentType(r, best)
	return best
}
