package songs

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:generate mockgen -package songs -typed -mock_names Service=MockSongService -destination ./mock_service_test.go github.com/Karaoke-Manager/karman/internal/service/song Service

func doRequest(t *testing.T, svc *MockSongService, req *http.Request) *http.Response {
	r := chi.NewRouter()
	r.Route("/", NewController(svc).Router)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Result()
}

func getMockService(t *testing.T) (*MockSongService, func()) {
	ctrl := gomock.NewController(t)
	return NewMockSongService(ctrl), ctrl.Finish
}
