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
		ctrl.Finish()
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
		ctrl.Finish()
		resp := w.Result()

		var err apierror.ProblemDetails
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, http.StatusNotFound, err.Status)
	})
}

func TestController_CheckModify(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		song := model.NewSong()
		song.UUID = id

		req := httptest.NewRequest(http.MethodPut, "/"+id.String(), nil)
		req = req.WithContext(SetSong(req.Context(), song))

		ctrl := gomock.NewController(t)
		handler := NewController(NewMockSongService(ctrl)).CheckModify(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		ctrl.Finish()
		resp := w.Result()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("fail", func(t *testing.T) {
		uploadID := uint(123)
		song := model.NewSong()
		song.UUID = id
		song.UploadID = &uploadID

		req := httptest.NewRequest(http.MethodPut, "/"+id.String(), nil)
		req = req.WithContext(SetSong(req.Context(), song))
		ctrl := gomock.NewController(t)
		handler := NewController(NewMockSongService(ctrl)).CheckModify(nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		ctrl.Finish()
		resp := w.Result()

		var err apierror.ProblemDetails
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&err))
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		assert.Equal(t, http.StatusConflict, err.Status)
		assert.Equal(t, apierror.TypeUploadSongReadonly, err.Type)
	})
}
