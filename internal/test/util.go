package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func MustOpen(t *testing.T, name string) *os.File {
	f, err := os.Open(name)
	require.NoErrorf(t, err, "could not open test file: %s", name)
	t.Cleanup(func() {
		require.NoErrorf(t, f.Close(), "could not close test file: %s", name)
	})
	return f
}

func DoRequest(h http.Handler, r *http.Request) *http.Response {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Result()
}

func AssertPagination(t *testing.T, resp *http.Response, offset, limit, count int, total int64) {
	assert.Equal(t, strconv.Itoa(offset), resp.Header.Get("Pagination-Offset"), "Pagination-Offset header does not match")
	assert.Equal(t, strconv.Itoa(limit), resp.Header.Get("Pagination-Limit"), "Pagination-Limit header does not match")
	assert.Equal(t, strconv.Itoa(count), resp.Header.Get("Pagination-Count"), "Pagination-Count header does not match")
	assert.Equal(t, strconv.FormatInt(total, 10), resp.Header.Get("Pagination-Total"), "Pagination-Total header does not match")
}
