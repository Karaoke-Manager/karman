package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/test"
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
			h := UUID("uuid")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if c.expectError {
					t.Errorf("UUID(%q) accepted %q, expected reject", "uuid", c.value)
					return
				}
				v, ok := GetUUID(r.Context())
				if !ok {
					t.Errorf("UUID(%q)(%q) did not set UUID in context, expected UUID to be set", "uuid", c.value)
				}
				if v != uuid.MustParse(c.value) {
					t.Errorf("UUID(%q)(%q) set UUID to %q, expected %q", "uuid", c.value, v, c.value)
				}
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", c.value), nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("uuid", c.value)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			resp := test.DoRequest(h, req)
			if !c.expectError {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("UUID(%q) rejected %q, expected accept", "uuid", c.value)
				}
				return
			}
			test.AssertProblemDetails(t, resp, http.StatusBadRequest, apierror.TypeInvalidUUID, nil)
		})
	}
}
