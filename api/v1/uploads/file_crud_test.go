package uploads

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Karaoke-Manager/karman/test"
)

func TestController_PutFile(t *testing.T) {
	h, _, data := setup(t, true)
	path := uploadPath(data.OpenUpload, "/files/foobar.txt")

	t.Run("204 No Content", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, path, strings.NewReader("Hello World"))
		r.Header.Set("Content-Type", "application/octet-stream")
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, "/"+data.InvalidUUID+"/files/foobar.txt"))
	t.Run("400 Bad Request (Bad Path)", testInvalidPath(h, http.MethodPut, uploadPath(data.OpenUpload, "/files/foo/../bar.txt"), data.OpenUpload.UUID, "foo/../bar.txt"))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, path, "application/octet-stream"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, uploadPath(data.AbsentUpload, "/files/anything.txt"), http.StatusNotFound))
	t.Run("409 Conflict", testInvalidState(h, http.MethodPut, uploadPath(data.ProcessingUpload, "/files/test.txt"), data.ProcessingUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, path, "video/mp4", "application/octet-stream"))
}
