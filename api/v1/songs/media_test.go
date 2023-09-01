//go:build database

package songs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"codello.dev/ultrastar/txt"
	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/test"
	testdata "github.com/Karaoke-Manager/karman/test/data"
)

func TestController_GetTxt(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	songWithCover := testdata.SongWithCover(t, db)
	url := fmt.Sprintf("/v1/songs/%s/txt", songWithCover.UUID)

	t.Run("200 OK", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, url, nil)
		resp := test.DoRequest(h, r)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s returned status %d, expected %d", url, resp.StatusCode, http.StatusOK)
		}
		if resp.Header.Get("Content-Disposition") == "" {
			t.Errorf("GET %s returned no Content-Disposition header, expected non-empty value", url)
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
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	songWithCover := testdata.SongWithCover(t, db)
	songWithUpload := testdata.SongWithUpload(t, db)
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
		if song.Title != "Foobar" {
			t.Errorf(`PUT %s responded with "#TITLE:%s", expected %q`, url, song.Title, "Foobar")
		}
		if !slices.Equal(song.Artists, []string{"Barfoo"}) {
			t.Errorf(`PUT %s responded with "#ARTIST:%v", expected %v`, url, song.Artists, []string{"Barfoo"})
		}
		if song.Cover == nil {
			t.Errorf(`PUT %s responded with no #COVER, expected non-empty string`, url)
		}
	})
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, fmt.Sprintf("/v1/songs/%s/txt", testdata.InvalidUUID)))
	t.Run("400 Bad Request (Invalid Body)", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, url, strings.NewReader(`Invalid Song`))
		r.Header.Set("Content-Type", "text/plain")
		resp := test.DoRequest(h, r)
		test.AssertProblemDetails(t, resp, http.StatusBadRequest, apierror.TypeInvalidTXT, map[string]any{
			"line": 1,
		})
	})
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, url, "text/plain", "text/x-ultrastar"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, fmt.Sprintf("/v1/songs/%s/txt", uuid.New()), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, "/v1/songs/%s/txt", songWithUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, url, "application/json", "text/plain", "text/x-ultrastar"))
}

func testMediaNotFound(h http.Handler, song model.Song, media string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/songs/%s/%s", song.UUID, media), nil)
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
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s %s responded with status code %d, expected %d", r.Method, r.RequestURI, resp.StatusCode, http.StatusOK)
		}
		if resp.Header.Get("Content-Disposition") == "" {
			t.Errorf("%s %s returned no Content-Disposition header, expected non-empty value", r.Method, r.RequestURI)
		}
		mType, err := mediatype.Parse(resp.Header.Get("Content-Type"))
		if err != nil {
			t.Errorf("%s %s responded with Content-Type:%s, expected valid media type: %s", r.Method, r.RequestURI, resp.Header.Get("Content-Type"), err)
		}
		if !mType.Equals(file.Type) {
			t.Errorf("%s %s responded with Content-Type:%s, expected %s", r.Method, r.RequestURI, mType, file.Type)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%s %s responded with an unreadable body: %s", r.Method, r.RequestURI, err)
			return
		}
		if int64(len(body)) != file.Size {
			t.Errorf("%s %s responded with %d bytes, expected %d", r.Method, r.RequestURI, len(body), file.Size)
		}
	}
}

func TestController_GetCover(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	c.mediaStore = media.NewMockStore("hello world")
	h := setupHandler(c, "/v1/songs/")
	simpleSong := testdata.SimpleSong(t, db)
	songWithCover := testdata.SongWithCover(t, db)
	songWithCover.CoverFile.Size = int64(len("hello world"))

	t.Run("200 OK", testGetFile(h, fmt.Sprintf("/v1/songs/%s/cover", songWithCover.UUID), *songWithCover.CoverFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s/cover", testdata.InvalidUUID)))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s/cover", uuid.New()), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, simpleSong, "cover"))
}

func TestController_GetBackground(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	c.mediaStore = media.NewMockStore("foobar")
	h := setupHandler(c, "/v1/songs/")
	simpleSong := testdata.SimpleSong(t, db)
	songWithBackground := testdata.SongWithBackground(t, db)
	songWithBackground.BackgroundFile.Size = int64(len("foobar"))

	t.Run("200 OK", testGetFile(h, fmt.Sprintf("/v1/songs/%s/background", songWithBackground.UUID), *songWithBackground.BackgroundFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s/background", testdata.InvalidUUID)))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s/background", uuid.New()), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, simpleSong, "background"))
}

func TestController_GetAudio(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	c.mediaStore = media.NewMockStore("some text")
	h := setupHandler(c, "/v1/songs/")
	simpleSong := testdata.SimpleSong(t, db)
	songWithAudio := testdata.SongWithAudio(t, db)
	songWithAudio.AudioFile.Size = int64(len("some text"))

	t.Run("200 OK", testGetFile(h, fmt.Sprintf("/v1/songs/%s/audio", songWithAudio.UUID), *songWithAudio.AudioFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s/audio", testdata.InvalidUUID)))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s/audio", uuid.New()), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, simpleSong, "audio"))
}

func TestController_GetVideo(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	c.mediaStore = media.NewMockStore("")
	h := setupHandler(c, "/v1/songs/")
	simpleSong := testdata.SimpleSong(t, db)
	songWithVideo := testdata.SongWithVideo(t, db)
	songWithVideo.VideoFile.Size = int64(len(""))

	t.Run("200 OK", testGetFile(h, fmt.Sprintf("/v1/songs/%s/video", songWithVideo.UUID), *songWithVideo.VideoFile))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s/video", testdata.InvalidUUID)))
	t.Run("404 Not Found (Song)", test.HTTPError(h, http.MethodGet, fmt.Sprintf("/v1/songs/%s/video", uuid.New()), http.StatusNotFound))
	t.Run("404 Not Found (Media)", testMediaNotFound(h, simpleSong, "video"))
}

func testPutFile(h http.Handler, path string, contentType string, content io.Reader) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, path, content)
		r.Header.Set("Content-Type", contentType)
		resp := test.DoRequest(h, r)
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("%s %s responded with status code %d, expected %d", r.Method, r.RequestURI, resp.StatusCode, http.StatusNoContent)
		}
	}
}

func TestController_ReplaceCover(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	c.mediaStore = media.NewMockStore("foobar")
	h := setupHandler(c, "/v1/songs/")
	songWithUpload := testdata.SongWithUpload(t, db)
	songWithCover := testdata.SongWithCover(t, db)
	url := fmt.Sprintf("/v1/songs/%s/cover", songWithCover.UUID)

	t.Run("204 No Content", testPutFile(h, url, "image/x-testing", strings.NewReader("foobar")))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, fmt.Sprintf("/v1/songs/%s/cover", testdata.InvalidUUID)))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, url, "image/*"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, fmt.Sprintf("/v1/songs/%s/cover", uuid.New()), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, "/v1/songs/%s/cover", songWithUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, url, "video/mp4", "image/*"))
}

func TestController_ReplaceBackground(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	c.mediaStore = media.NewMockStore("barfoo")
	h := setupHandler(c, "/v1/songs/")
	songWithUpload := testdata.SongWithUpload(t, db)
	songWithBackground := testdata.SongWithBackground(t, db)
	url := fmt.Sprintf("/v1/songs/%s/background", songWithBackground.UUID)

	t.Run("204 No Content", testPutFile(h, url, "image/x-testing", strings.NewReader("barfoo")))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, fmt.Sprintf("/v1/songs/%s/background", testdata.InvalidUUID)))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, url, "image/*"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, fmt.Sprintf("/v1/songs/%s/cover", uuid.New()), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, "/v1/songs/%s/background", songWithUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, url, "video/mp4", "image/*"))
}

func TestController_ReplaceAudio(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	c.mediaStore = media.NewMockStore("")
	h := setupHandler(c, "/v1/songs/")
	songWithUpload := testdata.SongWithUpload(t, db)
	songWithAudio := testdata.SongWithAudio(t, db)
	url := fmt.Sprintf("/v1/songs/%s/audio", songWithAudio.UUID)

	t.Run("204 No Content", testPutFile(h, url, "audio/x-testing", strings.NewReader("")))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, fmt.Sprintf("/v1/songs/%s/audio", testdata.InvalidUUID)))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, url, "audio/*"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, fmt.Sprintf("/v1/songs/%s/audio", uuid.New()), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, "/v1/songs/%s/audio", songWithUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, url, "video/mp4", "audio/*"))
}

func TestController_ReplaceVideo(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	c.mediaStore = media.NewMockStore("")
	h := setupHandler(c, "/v1/songs/")
	songWithUpload := testdata.SongWithUpload(t, db)
	songWithVideo := testdata.SongWithVideo(t, db)
	url := fmt.Sprintf("/v1/songs/%s/video", songWithVideo.UUID)

	t.Run("204 No Content", testPutFile(h, url, "video/x-testing", strings.NewReader("")))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodPut, fmt.Sprintf("/v1/songs/%s/video", testdata.InvalidUUID)))
	t.Run("400 Bad Request (Missing Content-Type)", test.MissingContentType(h, http.MethodPut, url, "video/*"))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodPut, fmt.Sprintf("/v1/songs/%s/video", uuid.New()), http.StatusNotFound))
	t.Run("409 Conflict", testSongConflict(h, http.MethodPut, "/v1/songs/%s/video", songWithUpload.UUID))
	t.Run("415 Unsupported Media Type", test.InvalidContentType(h, http.MethodPut, url, "audio/mp4", "video/*"))
}

func testDeleteFile(h http.Handler, path string) func(t *testing.T) {
	return func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, path, nil)
		resp := test.DoRequest(h, r)
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("%s %s responded with status code %d, expected %d", r.Method, r.RequestURI, resp.StatusCode, http.StatusNoContent)
		}

		// Repeat the request to test idempotency
		r = httptest.NewRequest(http.MethodDelete, path, nil)
		resp = test.DoRequest(h, r)
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("%s %s responded with status code %d, expected %d", r.Method, r.RequestURI, resp.StatusCode, http.StatusNoContent)
		}
	}
}

func TestController_DeleteCover(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	songWithCover := testdata.SongWithCover(t, db)

	t.Run("204 No Content", testDeleteFile(h, fmt.Sprintf("/v1/songs/%s/cover", songWithCover.UUID)))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodDelete, fmt.Sprintf("/v1/songs/%s/cover", testdata.InvalidUUID)))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodDelete, fmt.Sprintf("/v1/songs/%s/cover", uuid.New()), http.StatusNotFound))
}

func TestController_DeleteBackground(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	songWithBackground := testdata.SongWithBackground(t, db)

	t.Run("204 No Content", testDeleteFile(h, fmt.Sprintf("/v1/songs/%s/background", songWithBackground.UUID)))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodDelete, fmt.Sprintf("/v1/songs/%s/background", testdata.InvalidUUID)))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodDelete, fmt.Sprintf("/v1/songs/%s/background", uuid.New()), http.StatusNotFound))
}

func TestController_DeleteAudio(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	songWithAudio := testdata.SongWithAudio(t, db)

	t.Run("204 No Content", testDeleteFile(h, fmt.Sprintf("/v1/songs/%s/audio", songWithAudio.UUID)))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodDelete, fmt.Sprintf("/v1/songs/%s/audio", testdata.InvalidUUID)))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodDelete, fmt.Sprintf("/v1/songs/%s/audio", uuid.New()), http.StatusNotFound))
}

func TestController_DeleteVideo(t *testing.T) {
	t.Parallel()
	c, db := setupController(t)
	h := setupHandler(c, "/v1/songs/")
	songWithVideo := testdata.SongWithVideo(t, db)

	t.Run("204 No Content", testDeleteFile(h, fmt.Sprintf("/v1/songs/%s/video", songWithVideo.UUID)))
	t.Run("400 Bad Request (Invalid UUID)", test.InvalidUUID(h, http.MethodDelete, fmt.Sprintf("/v1/songs/%s/video", testdata.InvalidUUID)))
	t.Run("404 Not Found", test.HTTPError(h, http.MethodDelete, fmt.Sprintf("/v1/songs/%s/video", uuid.New()), http.StatusNotFound))
}
