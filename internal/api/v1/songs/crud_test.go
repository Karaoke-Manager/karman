package songs

import (
	"context"
	"encoding/json"
	"github.com/Karaoke-Manager/go-ultrastar"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strconv"
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
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
				req.Header.Set("Content-Type", c.mediaType)
				resp := doRequest(t, req, nil)

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
		expectedModel := model.Song{
			Title:      "Hello World",
			Artist:     "Foo",
			CalcMedley: true,
		}
		expectedSchema := schema.Song{
			SongRW: schema.SongRW{
				Title:  "Hello World",
				Artist: "Foo",
			},
		}
		expectedSchema.Medley.Mode = schema.MedleyModeAuto

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
		req.Header.Set("Content-Type", "text/plain")
		resp := doRequest(t, req, func(svc *MockSongService) {
			svc.EXPECT().CreateSong(gomock.Any(), expected).Return(expectedModel, nil)
		})

		var song schema.Song
		err := json.NewDecoder(resp.Body).Decode(&song)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, expectedSchema, song)
	})

	t.Run("syntax error", func(t *testing.T) {
		data := `#TITLE:Foo
unknown line`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
		req.Header.Set("Content-Type", "text/plain")
		resp := doRequest(t, req, nil)

		var err apierror.ProblemDetails
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, http.StatusBadRequest, err.Status)
		assert.Equal(t, apierror.TypeInvalidTXT, err.Type)
		assert.Equal(t, float64(2), err.Fields["line"])
	})
}

func TestController_Find(t *testing.T) {
	songs := make([]model.Song, 150)
	for i := range songs {
		songs[i] = model.Song{
			Title:  "Song " + strconv.Itoa(i),
			Artist: "Testing",
		}
	}
	cases := []struct {
		Name               string
		Default            bool
		Limit              string
		Offset             string
		ExpectRequestLimit int
		ExpectLimit        int
		ExpectOffset       int
		ExpectedCount      int
		ExpectErr          bool
	}{
		{"default", true, "0", "0", 25, 25, 0, 25, false},
		{"explicit limit", false, "10", "5", 10, 10, 5, 10, false},
		{"high limit", false, "130", "20", 130, 100, 20, 100, false},
		{"length past end", false, "50", "120", 50, 50, 120, 30, false},
		{"offset past end", false, "30", "170", 30, 30, 170, 0, false},
		{"negative values", false, "-25", "-3", -25, 0, 0, 0, false},
		{"invalid values", false, "foo", "bar", 0, 0, 0, 0, true},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if !c.Default {
				q := req.URL.Query()
				q.Add("limit", c.Limit)
				q.Add("offset", c.Offset)
				req.URL.RawQuery = q.Encode()
			}
			resp := doRequest(t, req, func(svc *MockSongService) {
				if c.ExpectErr {
					return
				}
				low := c.ExpectOffset
				if low > len(songs) {
					low = len(songs)
				}
				high := low + c.ExpectLimit
				if high > len(songs) {
					high = len(songs)
				}
				svc.EXPECT().FindSongs(gomock.Any(), c.ExpectLimit, c.ExpectOffset).Return(songs[low:high], int64(len(songs)), nil)
			})

			if c.ExpectErr {
				var err apierror.ProblemDetails
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				assert.Equal(t, http.StatusBadRequest, err.Status)
			} else {
				var page schema.List[*schema.Song]
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&page))
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.Len(t, page.Items, c.ExpectedCount)
				assert.Equal(t, c.ExpectRequestLimit, page.Limit)
				assert.Equal(t, c.ExpectedCount, page.Count)
				assert.Equal(t, int64(len(songs)), page.Total)
				assert.Equal(t, c.ExpectOffset, page.Offset)
			}
		})
	}
}

func TestController_Get(t *testing.T) {
	cases := []struct {
		name    string
		uuid    string
		success bool
	}{
		{"success", uuid.New().String(), true},
		{"fail", uuid.New().String(), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+c.uuid, nil)
			resp := doRequest(t, req, func(svc *MockSongService) {
				song := model.NewSong()
				if c.success {
					song.Title = "Foo"
					svc.EXPECT().GetSong(gomock.Any(), c.uuid).Return(song, nil)
				} else {
					svc.EXPECT().GetSong(gomock.Any(), c.uuid).Return(song, gorm.ErrRecordNotFound)
				}
			})
			if c.success {
				var song schema.Song
				assert.NoError(t, json.NewDecoder(resp.Body).Decode(&song))
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.Equal(t, "Foo", song.Title)
			} else {
				var err apierror.ProblemDetails
				assert.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
				assert.Equal(t, http.StatusNotFound, resp.StatusCode)
				assert.Equal(t, http.StatusNotFound, err.Status)
			}
		})
	}
}

func TestController_Update(t *testing.T) {
	id := uuid.New()
	uploadID := uint(123)
	song := model.NewSong()
	song.UUID = id
	songWithUpload := model.NewSong()
	songWithUpload.UUID = id
	songWithUpload.UploadID = &uploadID

	t.Run("simple", func(t *testing.T) {
		body := strings.NewReader(`{"title": "Hello World"}`)
		req := httptest.NewRequest(http.MethodPost, "/"+id.String(), body)
		req.Header.Set("Content-Type", "application/json")
		resp := doRequest(t, req, func(svc *MockSongService) {
			svc.EXPECT().GetSong(gomock.Any(), id.String()).Return(song, nil)
			svc.EXPECT().SaveSong(gomock.Any(), gomock.AssignableToTypeOf(&model.Song{})).DoAndReturn(func(ctx context.Context, song *model.Song) error {
				assert.Equal(t, id, song.UUID)
				assert.Equal(t, "Hello World", song.Title)
				return nil
			})
		})

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	cases := []struct {
		name        string
		body        string
		song        model.Song
		notFound    bool
		code        int
		problemType string
	}{
		{"not found", `{"title": "Hello `, song, true, http.StatusNotFound, ""},
		{"bad JSON", `{"title": "Hello `, song, false, http.StatusBadRequest, ""},
		{"schema validation", `{"title": "Hello World", "medley": {"mode": "manual"}}`, song, false, http.StatusUnprocessableEntity, ""},
		{"conflict", `{"title": "Hello World"}`, songWithUpload, false, http.StatusConflict, apierror.TypeUploadSongReadonly},
	}

	for _, c := range cases {
		body := strings.NewReader(c.body)
		req := httptest.NewRequest(http.MethodPost, "/"+c.song.UUID.String(), body)
		req.Header.Set("Content-Type", "application/json")
		resp := doRequest(t, req, func(svc *MockSongService) {
			if c.notFound {
				svc.EXPECT().GetSong(gomock.Any(), c.song.UUID.String()).Return(c.song, gorm.ErrRecordNotFound)
			} else {
				svc.EXPECT().GetSong(gomock.Any(), c.song.UUID.String()).Return(c.song, nil)
			}
		})

		var err apierror.ProblemDetails
		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
		assert.Equal(t, c.code, resp.StatusCode)
		assert.Equal(t, c.code, err.Status)
		if c.problemType == "" {
			assert.True(t, err.IsDefaultType())
		} else {
			assert.Equal(t, c.problemType, err.Type)
		}
	}
}
