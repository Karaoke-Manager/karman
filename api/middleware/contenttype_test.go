package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Karaoke-Manager/karman/api/apierror"
	_ "github.com/Karaoke-Manager/karman/pkg/render/json"
)

func TestRequireContentType(t *testing.T) {
	t.Run("no content types", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = RequireContentType()
		})
	})

	t.Run("invalid content type", func(t *testing.T) {
		cases := []string{"abc", "*", "/", ""}
		for _, c := range cases {
			t.Run(c, func(t *testing.T) {
				assert.Panics(t, func() {
					_ = RequireContentType(c)
				})
			})
		}
	})

	t.Run("fixed type", func(t *testing.T) {
		cases := map[string]struct {
			allowed string
			actual  string
			ok      bool
			code    int
		}{
			"not equal":            {"application/json", "foo/bar", false, http.StatusUnsupportedMediaType},
			"empty":                {"text/plain", "", false, http.StatusBadRequest},
			"star":                 {"image/png", "*", false, http.StatusUnsupportedMediaType},
			"wildcard":             {"application/json", "application/*", false, http.StatusUnsupportedMediaType},
			"correct fixed":        {"application/json", "application/json", true, http.StatusNoContent},
			"correct wildcard":     {"text/*", "text/plain", true, http.StatusNoContent},
			"incorrect wildcard":   {"text/*", "text/*", false, http.StatusUnsupportedMediaType},
			"empty wildcard match": {"text/*", "text/", false, http.StatusUnsupportedMediaType},
			"no slash":             {"text/*", "text", false, http.StatusUnsupportedMediaType},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				handler := RequireContentType(c.allowed)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.True(t, c.ok)
				}))
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Content-Type", c.actual)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				resp := w.Result()

				if !c.ok {
					var err apierror.ProblemDetails
					require.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
					assert.Equal(t, c.code, resp.StatusCode)
					assert.Equal(t, c.code, err.Status)
				}
			})
		}
	})
}
