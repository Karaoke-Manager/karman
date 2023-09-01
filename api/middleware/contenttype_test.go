package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Karaoke-Manager/karman/api/apierror"
	_ "github.com/Karaoke-Manager/karman/pkg/render/json"
	"github.com/Karaoke-Manager/karman/test"
)

func TestRequireContentType(t *testing.T) {
	t.Run("no content types", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("RequireContentType() did not panic, but a panic was expected.")
			}
		}()
		_ = RequireContentType()
	})

	t.Run("invalid content type", func(t *testing.T) {
		cases := []string{"abc", "*", "/", ""}
		for _, c := range cases {
			t.Run(c, func(t *testing.T) {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("RequireContentType(%q) did not panic, but a panic was expected.", c)
					}
				}()
				_ = RequireContentType(c)
			})
		}
	})

	t.Run("fixed type", func(t *testing.T) {
		cases := map[string]struct {
			allowed     string
			actual      string
			ok          bool
			code        int
			problemType string
		}{
			"not equal":            {"application/json", "foo/bar", false, http.StatusUnsupportedMediaType, apierror.TypeUnsupportedMediaType},
			"empty":                {"text/plain", "", false, http.StatusBadRequest, apierror.TypeMissingContentType},
			"star":                 {"image/png", "*", false, http.StatusUnsupportedMediaType, apierror.TypeUnsupportedMediaType},
			"wildcard":             {"application/json", "application/*", false, http.StatusUnsupportedMediaType, apierror.TypeUnsupportedMediaType},
			"correct fixed":        {"application/json", "application/json", true, http.StatusNoContent, ""},
			"correct wildcard":     {"text/*", "text/plain", true, http.StatusNoContent, ""},
			"incorrect wildcard":   {"text/*", "text/*", false, http.StatusUnsupportedMediaType, apierror.TypeUnsupportedMediaType},
			"empty wildcard match": {"text/*", "text/", false, http.StatusUnsupportedMediaType, apierror.TypeUnsupportedMediaType},
			"no slash":             {"text/*", "text", false, http.StatusUnsupportedMediaType, apierror.TypeUnsupportedMediaType},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				h := RequireContentType(c.allowed)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !c.ok {
						t.Errorf("RequireContentType(%q) accepted %q, expected reject", c.allowed, c.actual)
					}
					w.WriteHeader(c.code)
				}))
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Content-Type", c.actual)
				resp := test.DoRequest(h, req)
				if c.ok {
					if resp.StatusCode != c.code {
						t.Errorf("RequireContentType(%q) rejected %q, expected accept", c.allowed, c.actual)
					}
					return
				}

				test.AssertProblemDetails(t, resp, c.code, c.problemType, nil)
			})
		}
	})
}
