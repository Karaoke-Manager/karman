package songs

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/go-ultrastar/txt"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestController_GetTxt(t *testing.T) {
	id := uuid.New()
	song := model.NewSong()
	song.UUID = id
	usSong := ultrastar.NewSong()
	usSong.Title = "Foo"
	req := httptest.NewRequest(http.MethodGet, "/"+id.String()+"/txt", nil)
	resp := doRequest(t, req, func(svc *MockSongService, _ *MockMediaService) {
		svc.EXPECT().GetSongWithFiles(gomock.Any(), id).Return(song, nil)
		svc.EXPECT().SongData(song).Return(usSong)
	})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
	text, _ := io.ReadAll(resp.Body)
	expectedText := &bytes.Buffer{}
	_ = txt.WriteSong(expectedText, usSong)
	assert.Equal(t, expectedText.String(), string(text))
}

func TestController_ReplaceTxt(t *testing.T) {
	id := uuid.New()
	req := httptest.NewRequest(http.MethodPut, "/"+id.String()+"/txt", strings.NewReader("#TITLE:Foo"))
	req.Header.Set("Content-Type", "text/plain")
	resp := doRequest(t, req, func(svc *MockSongService, _ *MockMediaService) {
		s := model.NewSong()
		s.UUID = id
		svc.EXPECT().GetSong(gomock.Any(), id).Return(s, nil)
		svc.EXPECT().UpdateSongFromData(&s, gomock.Any()).DoAndReturn(func(song *model.Song, data *ultrastar.Song) {
			assert.Equal(t, "Foo", data.Title)
			song.Title = "Bar"
		})
		svc.EXPECT().SaveSong(gomock.Any(), gomock.Any()).Return(nil)
	})
	var s schema.Song
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&s))
	assert.Equal(t, "Bar", s.Title)
}

func TestController_GetMedia(t *testing.T) {
	id := uuid.New()
	song := model.NewSong()
	song.UUID = id
	song.CoverFile = &model.File{
		Type: "text/plain",
		Size: 1,
	}
	song.BackgroundFile = &model.File{
		Type: "text/plain",
		Size: 2,
	}
	song.AudioFile = &model.File{
		Type: "text/plain",
		Size: 3,
	}
	song.VideoFile = &model.File{
		Type: "text/plain",
		Size: 4,
	}
	songWithoutMedia := model.NewSong()
	songWithoutMedia.UUID = id

	cases := []struct {
		path  string
		field *model.File
	}{
		{"cover", song.CoverFile},
		{"background", song.BackgroundFile},
		{"audio", song.AudioFile},
		{"video", song.VideoFile},
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+id.String()+"/"+c.path, nil)
			resp := doRequest(t, req, func(songSvc *MockSongService, mediaSvc *MockMediaService) {
				songSvc.EXPECT().GetSongWithFiles(gomock.Any(), id).Return(song, nil)
				mediaSvc.EXPECT().ReadFile(gomock.Any(), *c.field).Return(io.NopCloser(strings.NewReader("content")), nil)
			})
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "content", string(body))
		})
		t.Run(c.path+" not found", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+id.String()+"/"+c.path, nil)
			resp := doRequest(t, req, func(songSvc *MockSongService, _ *MockMediaService) {
				songSvc.EXPECT().GetSongWithFiles(gomock.Any(), id).Return(songWithoutMedia, nil)
			})
			var err apierror.ProblemDetails
			assert.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
			assert.Equal(t, http.StatusNotFound, err.Status)
			assert.Equal(t, apierror.TypeMediaFileNotFound, err.Type)
			assert.Equal(t, id.String(), err.Fields["uuid"])
			assert.Equal(t, c.path, err.Fields["media"])
		})
	}
}

func TestController_ReplaceMedia(t *testing.T) {
	id := uuid.New()
	song := model.NewSong()
	song.UUID = id

	cases := []struct {
		path      string
		mediaType string
		image     bool
		field     **model.File
	}{
		{"cover", "image/png", true, &song.CoverFile},
		{"background", "image/jpeg", true, &song.BackgroundFile},
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/"+id.String()+"/"+c.path, strings.NewReader("content"))
			req.Header.Set("Content-Type", c.mediaType)
			resp := doRequest(t, req, func(songSvc *MockSongService, mediaSvc *MockMediaService) {
				songSvc.EXPECT().GetSong(gomock.Any(), id).Return(song, nil)
				file := model.File{}
				file.ID = 123
				file.UUID = uuid.New()
				contentHandler := func(ctx context.Context, _ string, r io.Reader) (model.File, error) {
					data, err := io.ReadAll(r)
					assert.NoError(t, err)
					assert.Equal(t, "content", string(data))
					return file, nil
				}
				if c.image {
					mediaSvc.EXPECT().StoreImageFile(gomock.Any(), c.mediaType, gomock.Any()).DoAndReturn(contentHandler)
				}
				*c.field = &file
				songSvc.EXPECT().SaveSong(gomock.Any(), &song).Return(nil)
			})
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		})
	}

}

func TestController_DeleteMedia(t *testing.T) {
	id := uuid.New()
	song := model.NewSong()
	song.UUID = id

	cases := []struct {
		path  string
		field **uint
	}{
		{"cover", &song.CoverFileID},
		{"background", &song.BackgroundFileID},
		{"audio", &song.AudioFileID},
		{"video", &song.VideoFileID},
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/"+id.String()+"/"+c.path, nil)
			resp := doRequest(t, req, func(songSvc *MockSongService, _ *MockMediaService) {
				cid := uint(123)
				*c.field = &cid
				songSvc.EXPECT().GetSong(gomock.Any(), id).Return(song, nil)
				*c.field = nil
				songSvc.EXPECT().SaveSong(gomock.Any(), &song).Return(nil)
			})
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		})
	}
}
