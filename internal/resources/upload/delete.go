package upload

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (s *Server) Delete(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "id")
}
