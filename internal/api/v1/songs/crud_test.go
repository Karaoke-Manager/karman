package songs

import (
	"encoding/json"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestController_Create(t *testing.T) {
	t.Run("content type", func(t *testing.T) {
		cases := map[string]struct {
			mediaType string
			code      int
		}{
			"invalid":  {"foo/bar", http.StatusUnsupportedMediaType},
			"empty":    {"", http.StatusBadRequest},
			"star":     {"*", http.StatusUnsupportedMediaType},
			"wildcard": {"text/*", http.StatusUnsupportedMediaType},
		}
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				svc, assertServiceCalls := getMockService(t)
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
				req.Header.Set("Content-Type", c.mediaType)
				resp := doRequest(t, svc, req)
				assertServiceCalls()

				var err apierror.ProblemDetails
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
				assert.Equal(t, c.code, resp.StatusCode)
				assert.Equal(t, c.code, err.Status)
			})
		}
	})

	t.Run("simple", func(t *testing.T) {
		data := `#TITLE:Hello World
#ARTIST:Foo
#BPM:12`
		expected := ultrastar.NewSongWithBPM(12 * 4)
		expected.Title = "Hello World"
		expected.Artist = "Foo"
		expectedModel := model.NewSongWithData(expected)
		expectedSchema := schema.NewSongFromModel(expectedModel)
		expectedSchema.Extra = nil

		svc, assertServiceCalls := getMockService(t)
		svc.EXPECT().CreateSong(gomock.Any(), expected).Return(expectedModel, nil)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
		req.Header.Set("Content-Type", "text/plain")
		resp := doRequest(t, svc, req)
		assertServiceCalls()

		var song schema.Song
		err := json.NewDecoder(resp.Body).Decode(&song)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, expectedSchema, song)
	})

	t.Run("syntax error", func(t *testing.T) {
		data := `#TITLE:Foo
unknown line`
		svc, assertServiceCalls := getMockService(t)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
		req.Header.Set("Content-Type", "text/plain")
		resp := doRequest(t, svc, req)
		assertServiceCalls()

		var err apierror.ProblemDetails
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, http.StatusBadRequest, err.Status)
		assert.Equal(t, apierror.TypeInvalidTXT, err.Type)
		assert.Equal(t, float64(2), err.Fields["line"])
	})
}
