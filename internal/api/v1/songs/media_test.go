package songs

import (
	"bytes"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/go-ultrastar/txt"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
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
