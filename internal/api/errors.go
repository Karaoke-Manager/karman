package api

import (
	"github.com/Karaoke-Manager/karman/internal/apierror"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"net/http"
)

func (s *Server) NotFound(w http.ResponseWriter, r *http.Request) {
	_ = render.Render(w, r, apierror.HttpStatus(http.StatusNotFound))
}

func (s *Server) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	// TODO: Include Allow Header: https://github.com/go-chi/chi/issues/446
	_ = render.Render(w, r, apierror.HttpStatus(http.StatusMethodNotAllowed))
}
