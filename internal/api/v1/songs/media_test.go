package songs

import (
	"encoding/json"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/Karaoke-Manager/karman/internal/test"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestController_GetTxt(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, songPath(data.SongWithCover, "/txt"), nil)
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, `#TITLE:Some
#ARTIST:Unimportant
#COVER:Unimportant - Some [CO].png
E
`, string(body))
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/txt"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodGet, songPath(data.AbsentSong, "/txt"), http.StatusNotFound))
}

func TestController_ReplaceTxt(t *testing.T) {
	h, _, data := setup(t, true)
	path := songPath(data.SongWithCover, "/txt")

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, path, strings.NewReader(`#TITLE:Foobar
#ARTIST:Barfoo`))
		r.Header.Set("Content-Type", "text/plain")
		resp := test.DoRequest(h, r)
		var song schema.Song
		if assert.NoError(t, json.NewDecoder(resp.Body).Decode(&song)) {
			assert.Equal(t, "Foobar", song.Title)
			assert.Equal(t, "Barfoo", song.Artist)
			assert.NotNil(t, song.Cover)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, "/"+data.InvalidUUID+"/txt"))
	t.Run("400 Bad Request (Invalid Body)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, path, strings.NewReader(`Invalid Song`))
		r.Header.Set("Content-Type", "text/plain")
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusBadRequest, apierror.TypeInvalidTXT, map[string]any{
			"line": 1,
		})
	})
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, path, "text/plain"))
	t.Run("400 Bad Request (Invalid Content-Type)", test.InvalidContentType(h, http.MethodPut, path, "application/json", "text/plain"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, songPath(data.AbsentSong, "/txt"), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, songPath(data.SongWithUpload, "/txt"), data.SongWithUpload.UUID))
}

func testMediaNotFound(h http.Handler, song model.Song, media string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, songPath(song, "/"+media), nil)
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusNotFound, apierror.TypeMediaFileNotFound, map[string]any{
			"uuid":  song.UUID.String(),
			"media": media,
		})
	}
}

func testGetFile(h http.Handler, path string, file model.File) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, strconv.FormatInt(file.Size, 10), resp.Header.Get("Content-Length"))
		assert.Equal(t, file.Type, resp.Header.Get("Content-Type"))
	}
}

func TestController_GetCover(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", testGetFile(h, songPath(data.SongWithCover, "/cover"), data.ImageFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/cover"))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, songPath(data.AbsentSong, "/cover"), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, data.SongWithoutMediaAndMusic, "cover"))
}

func TestController_GetBackground(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", testGetFile(h, songPath(data.SongWithBackground, "/background"), data.ImageFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/background"))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, songPath(data.AbsentSong, "/background"), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, data.SongWithoutMediaAndMusic, "background"))
}
func TestController_GetAudio(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", testGetFile(h, songPath(data.SongWithAudio, "/audio"), data.AudioFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/audio"))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, songPath(data.AbsentSong, "/audio"), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, data.SongWithoutMediaAndMusic, "audio"))
}
func TestController_GetVideo(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", testGetFile(h, songPath(data.SongWithVideo, "/video"), data.VideoFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/video"))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, songPath(data.AbsentSong, "/video"), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, data.SongWithoutMediaAndMusic, "video"))
}

func testPutFile(h http.Handler, path string, contentType string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, path, strings.NewReader("some content"))
		r.Header.Set("Content-Type", contentType)
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	}
}

func TestController_ReplaceCover(t *testing.T) {
	h, _, data := setup(t, true)
	path := songPath(data.SongWithCover, "/cover")

	t.Run("204 No Content", testPutFile(h, path, "image/x-testing"))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, "/"+data.InvalidUUID+"/cover"))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, path, "image/*"))
	t.Run("400 Bad Request (Invalid Content-Type)", test.InvalidContentType(h, http.MethodPut, path, "video/mp4", "image/*"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, songPath(data.AbsentSong, "/cover"), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, songPath(data.SongWithUpload, "/cover"), data.SongWithUpload.UUID))
}

func TestController_ReplaceBackground(t *testing.T) {
	h, _, data := setup(t, true)
	path := songPath(data.SongWithBackground, "/background")

	t.Run("204 No Content", testPutFile(h, path, "image/x-testing"))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, "/"+data.InvalidUUID+"/background"))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, path, "image/*"))
	t.Run("400 Bad Request (Invalid Content-Type)", test.InvalidContentType(h, http.MethodPut, path, "video/mp4", "image/*"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, songPath(data.AbsentSong, "/background"), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, songPath(data.SongWithUpload, "/background"), data.SongWithUpload.UUID))
}

func testDeleteFile(h http.Handler, path string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, path, nil)
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Repeat the request to test idempotency
		r = httptest.NewRequest(http.MethodDelete, path, nil)
		resp = test.DoRequest(h, r)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	}
}

func TestController_DeleteCover(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("204 No Content", testDeleteFile(h, songPath(data.SongWithCover, "/cover")))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodDelete, "/"+data.InvalidUUID+"/cover"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodDelete, songPath(data.AbsentSong, "/cover"), http.StatusNotFound))
}

func TestController_DeleteBackground(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("204 No Content", testDeleteFile(h, songPath(data.SongWithBackground, "/background")))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodDelete, "/"+data.InvalidUUID+"/background"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodDelete, songPath(data.AbsentSong, "/background"), http.StatusNotFound))
}

func TestController_DeleteAudio(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("204 No Content", testDeleteFile(h, songPath(data.SongWithAudio, "/audio")))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodDelete, "/"+data.InvalidUUID+"/audio"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodDelete, songPath(data.AbsentSong, "/audio"), http.StatusNotFound))
}

func TestController_DeleteVideo(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("204 No Content", testDeleteFile(h, songPath(data.SongWithVideo, "/video")))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodDelete, "/"+data.InvalidUUID+"/video"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodDelete, songPath(data.AbsentSong, "/video"), http.StatusNotFound))
}
