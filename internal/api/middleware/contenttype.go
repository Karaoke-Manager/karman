package middleware

import (
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"mime"
	"net/http"
	"strings"
)

type contentTypeMap map[string][]string

func (m contentTypeMap) Add(key, value string) {
	key = strings.ToLower(key)
	m[key] = append(m[key], strings.ToLower(value))
}

func (m contentTypeMap) Set(key, value string) {
	m[strings.ToLower(key)] = []string{value}
}

func (m contentTypeMap) Get(key string) string {
	if m == nil {
		return ""
	}
	v := m[strings.ToLower(key)]
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func (m contentTypeMap) Values(key string) []string {
	if m == nil {
		return nil
	}
	return m[strings.ToLower(key)]
}

func (m contentTypeMap) Del(key string) {
	delete(m, strings.ToLower(key))
}

func RequireContentType(contentTypes ...string) func(next http.Handler) http.Handler {
	allowed := contentTypeMap{}
	for i := range contentTypes {
		t, _, err := mime.ParseMediaType(contentTypes[i])
		if err != nil {
			panic("invalid media type: " + contentTypes[i])
		}
		media, sub, ok := strings.Cut(t, "/")
		if !ok {
			panic("invalid media type: " + contentTypes[i])
		}
		allowed.Add(media, sub)
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			s, _, err := mime.ParseMediaType(strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type"))))
			if err != nil {
				_ = render.Render(w, r, apierror.MissingContentType(contentTypes...))
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
			allowedSubs := allowed.Values(media)
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
