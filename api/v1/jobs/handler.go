package jobs

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/Karaoke-Manager/karman/task"
)

type Handler struct {
	logger *slog.Logger
	r      chi.Router

	cronService task.CronService
}

func NewHandler(
	logger *slog.Logger,
	cronService task.CronService,
) *Handler {
	r := chi.NewRouter()
	h := &Handler{
		logger,
		r,
		cronService,
	}

	r.With(render.ContentTypeNegotiation("application/json")).Get("/", h.List)
	r.With(render.ContentTypeNegotiation("application/json")).Get("/{name}", h.Get)
	r.Post("/{name}/start", h.Start)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}
