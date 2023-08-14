package middleware

import (
	"net/http"

	"github.com/Karaoke-Manager/server/internal/api/apierror"
	"github.com/Karaoke-Manager/server/pkg/mediatype"
	"github.com/Karaoke-Manager/server/pkg/render"
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
func RequireContentType(types ...string) func(next http.Handler) http.Handler {
	if len(types) == 0 {
		panic("no media types specified")
	}
	allowed := make(mediatype.MediaTypes, len(types))
	for i, t := range types {
		allowed[i] = mediatype.MustParse(t)
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("Content-Type")
			if raw == "" {
				_ = render.Render(w, r, apierror.MissingContentType(allowed...))
				return
			}
			t, err := mediatype.Parse(raw)
			if err != nil {
				// FIXME: Bad Request??
				_ = render.Render(w, r, apierror.UnsupportedMediaType(allowed...))
				return
			}
			if !t.IsConcrete() || !allowed.Includes(t) {
				// FIXME: Maybe Bad Request for non-concrete types?
				_ = render.Render(w, r, apierror.UnsupportedMediaType(allowed...))
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
