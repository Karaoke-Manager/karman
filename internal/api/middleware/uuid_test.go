package middleware

import (
	"context"
	"encoding/json"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUUID(t *testing.T) {
	cases := map[string]struct {
		value       string
		expectError bool
	}{
		"valid":   {"A37FCD49-40A2-4FB4-83AA-49A57B62317F", false},
		"invalid": {"Hello%20World", true},
		"empty":   {"", true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			handler := UUID("uuid")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				v := MustGetUUID(r.Context())
				assert.Equal(t, uuid.MustParse(c.value), v)
			}))
			req := httptest.NewRequest(http.MethodGet, "/"+c.value, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("uuid", c.value)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			resp := w.Result()

			if c.expectError {
				var err apierror.ProblemDetails
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				assert.Equal(t, http.StatusBadRequest, err.Status)
				assert.Equal(t, apierror.TypeInvalidUUID, err.Type)
			}
		})
	}
}
