package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

// MustOpen opens the named file.
// The file is closed automatically at the end of the test.
// If the file cannot be opened the test is aborted.
func MustOpen(t *testing.T, name string) *os.File {
	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("MustOpen() could not open test file %s: %s", name, err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Fatalf("MustOpen() could not close test file %s: %s", name, err)
		}
	})
	return f
}

// DoRequest executes r against h, records the response and returns it.
func DoRequest(h http.Handler, r *http.Request) *http.Response {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	resp := w.Result()
	resp.Request = r
	return resp
}

// AssertPagination validates that the response provides the expected pagination values.
func AssertPagination(t *testing.T, resp *http.Response, offset, limit, count int, total int64) {
	if resp.Header.Get("Pagination-Offset") != strconv.Itoa(offset) {
		t.Errorf("%s %s responded with Pagination-Offset:%s, expected %d", resp.Request.Method, resp.Request.RequestURI, resp.Header.Get("Pagination-Offset"), offset)
	}
	if resp.Header.Get("Pagination-Limit") != strconv.Itoa(limit) {
		t.Errorf("%s %s responded with Pagination-Limit:%s, expected %d", resp.Request.Method, resp.Request.RequestURI, resp.Header.Get("Pagination-Limit"), limit)
	}
	if resp.Header.Get("Pagination-Count") != strconv.Itoa(count) {
		t.Errorf("%s %s responded with Pagination-Count:%s, expected %d", resp.Request.Method, resp.Request.RequestURI, resp.Header.Get("Pagination-Count"), count)
	}
	if resp.Header.Get("Pagination-Total") != strconv.FormatInt(total, 10) {
		t.Errorf("%s %s responded with Pagination-Total:%s, expected %d", resp.Request.Method, resp.Request.RequestURI, resp.Header.Get("Pagination-Total"), total)
	}
}
