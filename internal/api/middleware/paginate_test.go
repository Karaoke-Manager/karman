package middleware

import (
	"encoding/json"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaginate(t *testing.T) {
	cases := map[string]struct {
		maxLimit             int
		defaultLimit         int
		noReqLimit           bool
		reqLimit             string
		noReqOffset          bool
		reqOffset            string
		expErr               bool
		expectedRequestLimit int
		expectedLimit        int
		expectedOffset       int
	}{
		"standard":        {100, 25, false, "10", false, "0", false, 10, 10, 0},
		"default":         {100, 25, true, "", true, "", false, 25, 25, 0},
		"invalid":         {100, 25, false, "foo", false, "bar", true, 0, 0, 0},
		"over max":        {100, 25, false, "1000", false, "5", false, 1000, 100, 5},
		"large offset":    {100, 25, false, "100", false, "1000", false, 100, 100, 1000},
		"negative values": {100, 25, false, "-10", false, "-2", false, -10, 0, 0},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			handler := Paginate(c.defaultLimit, c.maxLimit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p := MustGetPagination(r.Context())
				assert.False(t, c.expErr)
				assert.Equal(t, c.expectedRequestLimit, p.RequestLimit)
				assert.Equal(t, c.expectedLimit, p.Limit)
				assert.Equal(t, c.expectedOffset, p.Offset)
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			q := req.URL.Query()
			if !c.noReqLimit {
				q.Add("limit", c.reqLimit)
			}
			if !c.noReqOffset {
				q.Add("offset", c.reqOffset)
			}
			req.URL.RawQuery = q.Encode()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			resp := w.Result()

			if c.expErr {
				var err apierror.ProblemDetails
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				assert.Equal(t, http.StatusBadRequest, err.Status)
			}
		})
	}
}
