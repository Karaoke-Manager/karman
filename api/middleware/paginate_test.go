package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Karaoke-Manager/karman/test"
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
		expectedOffset       int64
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
			h := Paginate(c.defaultLimit, c.maxLimit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p := MustGetPagination(r.Context())
				if c.expErr {
					t.Errorf("Paginate(%d, %d) accepted limit=%q, offset=%q, expected reject", c.maxLimit, c.defaultLimit, c.reqLimit, c.reqOffset)
				}
				if p.RequestLimit != c.expectedRequestLimit {
					t.Errorf("Paginate(%d, %d)(limit=%q, offset=%q) yielded p.RequestLimit=%d, expected %d", c.maxLimit, c.defaultLimit, c.reqLimit, c.reqOffset, p.RequestLimit, c.expectedRequestLimit)
				}
				if p.Limit != c.expectedLimit {
					t.Errorf("Paginate(%d, %d)(limit=%q, offset=%q) yielded p.Limit=%d, expected %d", c.maxLimit, c.defaultLimit, c.reqLimit, c.reqOffset, p.Limit, c.expectedLimit)
				}
				if p.Offset != c.expectedOffset {
					t.Errorf("Paginate(%d, %d)(limit=%q, offset=%q) yielded p.Offset=%d, expected %d", c.maxLimit, c.defaultLimit, c.reqLimit, c.reqOffset, p.Offset, c.expectedOffset)
				}
				w.WriteHeader(http.StatusOK)
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
			req.RequestURI = "/?" + req.URL.RawQuery
			resp := test.DoRequest(h, req) //nolint:bodyclose
			if !c.expErr {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Paginate(%d, %d) rejected limit=%q, offset=%q, expected accept", c.maxLimit, c.defaultLimit, c.reqLimit, c.reqOffset)
				}
				return
			}
			test.AssertProblemDetails(t, resp, http.StatusBadRequest, "", nil)
		})
	}
}
