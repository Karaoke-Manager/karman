package songs

import (
	"encoding/json"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/api/middleware"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestController_FetchUpload(t *testing.T) {
	id := uuid.New()
	song := model.NewSong()
	song.UUID = id

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/"+id.String(), nil)
		req = req.WithContext(middleware.SetUUID(req.Context(), id))
		ctrl := gomock.NewController(t)
		svc := NewMockSongService(ctrl)
		c := NewController(svc)
		svc.EXPECT().GetSong(gomock.Any(), id).Return(song, nil)
		handler := c.FetchUpload(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s := MustGetSong(r.Context())
			assert.Equal(t, id, s.UUID)
		}))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/"+id.String(), nil)
		req = req.WithContext(middleware.SetUUID(req.Context(), id))
		ctrl := gomock.NewController(t)
		svc := NewMockSongService(ctrl)
		c := NewController(svc)
		svc.EXPECT().GetSong(gomock.Any(), id).Return(model.Song{}, gorm.ErrRecordNotFound)
		handler := c.FetchUpload(false)(nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		resp := w.Result()

		var err apierror.ProblemDetails
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, http.StatusNotFound, err.Status)
	})
}
