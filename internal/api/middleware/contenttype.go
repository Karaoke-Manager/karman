package middleware

import (
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"mime"
	"net/http"
	"strings"
)

// ContentTypeJSON is an instance of the RequireContentType middleware for the common application of JSON requests.
var ContentTypeJSON = RequireContentType("application/json")

// RequireContentType is a middleware that enforces the use of the Content-Type header.
// To create this middleware you need to pass the allowed media types.
// You can pass parameters for media types, but those are ignored in this middleware.
//
// In addition to static media types like text/plain you can also use
// wildcard media types like text/* or the special value */* to accept all media types.
// Wildcard types like text/* are matched as a prefix of the request header (but parameters are ignored during matching).
// The special value */* allows all media type but still requires a Content-Type header to be present.
// In this case the syntax of the Content-Type header is not validated.
//
// Using this middleware forces the use of the Content-Type header.
// A request without this header will result in an error, even if the request body is empty.
func RequireContentType(mediaTypes ...string) func(next http.Handler) http.Handler {
	if len(mediaTypes) == 0 {
		panic("no media types specified")
	}
	allowed := map[string][]string{}
	allowAll := false
	for i := range mediaTypes {
		t, _, err := mime.ParseMediaType(mediaTypes[i])
		if err != nil {
			panic("invalid media type: " + mediaTypes[i])
		}
		media, sub, ok := strings.Cut(t, "/")
		if !ok {
			panic("invalid media type: " + mediaTypes[i])
		}
		if media == "" || sub == "" {
			panic("invalid media type: " + mediaTypes[i])
		}
		if media == "*" && sub == "*" {
			allowAll = true
		}
		if media == "*" && sub != "*" {
			panic("media types other than */* cannot start with a wildcard.")
		}
		allowed[media] = append(allowed[media], sub)
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
			if contentType == "" {
				_ = render.Render(w, r, apierror.MissingContentType(mediaTypes...))
				return
			}
			s, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				_ = render.Render(w, r, apierror.ErrUnsupportedMediaType)
				return
			}
			if allowAll {
				// Shortcut if all content types are accepted.
				// In this case we do not validate the format of the content type request header.
				next.ServeHTTP(w, r)
				return
			}
			media, sub, ok := strings.Cut(s, "/")
			if !ok {
				_ = render.Render(w, r, apierror.ErrUnsupportedMediaType)
				return
			}
			sub = strings.ToLower(sub)
			if sub == "*" {
				// FIXME: Bad Request?
				_ = render.Render(w, r, apierror.ErrUnsupportedMediaType)
				return
			}
			allowedSubs := allowed[media]
			for _, allowedSub := range allowedSubs {
				if allowedSub == "*" || allowedSub == sub {
					next.ServeHTTP(w, r)
					return
				}
			}
			_ = render.Render(w, r, apierror.ErrUnsupportedMediaType)
		}
		return http.HandlerFunc(fn)
	}
}
