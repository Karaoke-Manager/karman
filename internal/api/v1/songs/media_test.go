package songs

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/go-ultrastar/txt"
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
	resp := doRequest(t, req, func(svc *MockSongService) {
		svc.EXPECT().GetSongWithFiles(gomock.Any(), id).Return(song, nil)
		svc.EXPECT().UltraStarSong(gomock.Any(), song).Return(usSong)
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
	song := model.NewSong()
	song.UUID = id
	req := httptest.NewRequest(http.MethodPut, "/"+id.String()+"/txt", strings.NewReader("#TITLE:Foo"))
	req.Header.Set("Content-Type", "text/plain")
	resp := doRequest(t, req, func(svc *MockSongService) {
		svc.EXPECT().GetSongWithFiles(gomock.Any(), id).Return(song, nil)
		svc.EXPECT().ReplaceSong(gomock.Any(), &song, gomock.Any()).DoAndReturn(func(ctx context.Context, song *model.Song, data *ultrastar.Song) error {
			assert.Equal(t, "Foo", data.Title)
			song.Title = "Bar"
			return nil
		})
	})
	var s schema.Song
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&s))
	assert.Equal(t, "Bar", s.Title)
}
