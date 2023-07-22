package songs

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:generate mockgen -package songs -typed -mock_names Service=MockSongService -destination ./mock_service_test.go github.com/Karaoke-Manager/karman/internal/service/song Service

// doRequest executes the specified request against a songs.Controller backed by a MockSongService.
// Before the request is executed the expect function is invoked giving you the opportunity to register expected calls on the service.
func doRequest(t *testing.T, req *http.Request, expect func(svc *MockSongService)) *http.Response {
	ctrl := gomock.NewController(t)
	svc := NewMockSongService(ctrl)
	if expect != nil {
		expect(svc)
	}
	r := chi.NewRouter()
	r.Route("/", NewController(svc).Router)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	ctrl.Finish()
	return w.Result()
}
