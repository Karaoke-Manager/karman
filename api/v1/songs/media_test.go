//go:build database

package songs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"codello.dev/ultrastar/txt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func TestController_GetTxt(t *testing.T) {
	h, db := setupHandler(t, "/v1/songs/")
	songWithCover := testdata.SongWithCover(t, db)
	url := fmt.Sprintf("/v1/songs/%s/txt", songWithCover.UUID)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET /{uuid}/txt returned status %d, expected %d", resp.StatusCode, http.StatusOK)
		}
		body, err := txt.NewReader(resp.Body).ReadSong()
		if err != nil {
			t.Errorf("GET %s responded with an invalid UltraStar song", url)
		}
		if body.Title != songWithCover.Title {
			t.Errorf(`GET %s responded with "#TITLE:%s", expected %q`, url, body.Title, songWithCover.Title)
		}
		if body.CoverFileName == "" {
			t.Errorf("GET %s responded with no #COVER, expected non-empty string", url)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s/txt", testdata.InvalidUUID)))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s/txt", uuid.New()), http.StatusNotFound))
}

func TestController_ReplaceTxt(t *testing.T) {
	h, db := setupHandler(t, "/v1/songs/")
	songWithCover := testdata.SongWithCover(t, db)
	url := fmt.Sprintf("/v1/songs/%s/txt", songWithCover.UUID)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, url, strings.NewReader(`#TITLE:Foobar
#ARTIST:Barfoo`))
		r.Header.Set("Content-Type", "text/plain")
		resp := test.DoRequest(h, r)

		var song schema.Song
		if resp.StatusCode != http.StatusOK {
			t.Errorf("PUT %s responded with status code %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if err := json.NewDecoder(resp.Body).Decode(&song); err != nil {
			t.Errorf("PUT %s responded invalid song schema: %s", url, err)
			return
		}
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
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, path, "text/plain", "text/x-ultrastar"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, songPath(data.AbsentSong, "/txt"), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, songPath(data.SongWithUpload, "/txt"), data.SongWithUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, path, "application/json", "text/plain", "text/x-ultrastar"))
}

func testMediaNotFound(h http.Handler, song *model.Song, media string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, songPath(song, "/"+media), nil)
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusNotFound, apierror.TypeMediaFileNotFound, map[string]any{
			"uuid":  song.UUID.String(),
			"media": media,
		})
	}
}

func testGetFile(h http.Handler, path string, file *model.File) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		resp := test.DoRequest(h, r)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, strconv.FormatInt(file.Size, 10), resp.Header.Get("Content-Length"))
		assert.Equal(t, file.Type, mediatype.MustParse(resp.Header.Get("Content-Type")))
	}
}

func TestController_GetCover(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", testGetFile(h, songPath(data.SongWithCover, "/cover"), data.ImageFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/cover"))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, songPath(data.AbsentSong, "/cover"), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, data.BasicSong, "cover"))
}

func TestController_GetBackground(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", testGetFile(h, songPath(data.SongWithBackground, "/background"), data.ImageFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/background"))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, songPath(data.AbsentSong, "/background"), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, data.BasicSong, "background"))
}

func TestController_GetAudio(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", testGetFile(h, songPath(data.SongWithAudio, "/audio"), data.AudioFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/audio"))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, songPath(data.AbsentSong, "/audio"), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, data.BasicSong, "audio"))
}

func TestController_GetVideo(t *testing.T) {
	h, _, data := setup(t, true)

	t.Run("200 OK", testGetFile(h, songPath(data.SongWithVideo, "/video"), data.VideoFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, "/"+data.InvalidUUID+"/video"))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, songPath(data.AbsentSong, "/video"), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, data.BasicSong, "video"))
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
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, songPath(data.AbsentSong, "/cover"), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, songPath(data.SongWithUpload, "/cover"), data.SongWithUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, path, "video/mp4", "image/*"))
}

func TestController_ReplaceBackground(t *testing.T) {
	h, _, data := setup(t, true)
	path := songPath(data.SongWithBackground, "/background")

	t.Run("204 No Content", testPutFile(h, path, "image/x-testing"))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, "/"+data.InvalidUUID+"/background"))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, path, "image/*"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, songPath(data.AbsentSong, "/background"), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, songPath(data.SongWithUpload, "/background"), data.SongWithUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, path, "video/mp4", "image/*"))
}

func TestController_ReplaceAudio(t *testing.T) {
	h, _, data := setup(t, true)
	path := songPath(data.SongWithAudio, "/audio")

	t.Run("204 No Content", testPutFile(h, path, "audio/x-testing"))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, "/"+data.InvalidUUID+"/audio"))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, path, "audio/*"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, songPath(data.AbsentSong, "/audio"), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, songPath(data.SongWithUpload, "/audio"), data.SongWithUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, path, "video/mp4", "audio/*"))
}

func TestController_ReplaceVideo(t *testing.T) {
	h, _, data := setup(t, true)
	path := songPath(data.SongWithVideo, "/video")

	t.Run("204 No Content", testPutFile(h, path, "video/x-testing"))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, "/"+data.InvalidUUID+"/video"))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, path, "video/*"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, songPath(data.AbsentSong, "/video"), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, songPath(data.SongWithUpload, "/video"), data.SongWithUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, path, "audio/mp4", "video/*"))
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
