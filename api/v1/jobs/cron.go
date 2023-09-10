package jobs

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lmittmann/tint"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/Karaoke-Manager/karman/task"
)

// List implements the GET /v1/jobs endpoint.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	stats, err := h.cronService.ListJobs()
	if err != nil {
		h.logger.Error("Could not list jobs.", tint.Err(err))
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	result := make(map[string]schema.Job, len(stats))
	for name, stat := range stats {
		result[name] = schema.FromJobStat(stat)
	}
	// TODO: Use Render instead
	_ = render.Respond(w, r, result)
}

// Get implements the GET /v1/jobs/{name} endpoint.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	stat, err := h.cronService.StatJob(name)
	if errors.Is(err, core.ErrNotFound) {
		_ = render.Render(w, r, apierror.ErrNotFound)
		return
	} else if err != nil {
		h.logger.Error("Could not fetch job.", "name", name, tint.Err(err))
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	job := schema.FromJobStat(stat)
	// TODO: Use Render instead
	_ = render.Respond(w, r, job)
}

// Start implements the GET /v1/jobs/{name}/start endpoint.
func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	err := h.cronService.RunJob(name)
	if errors.Is(err, core.ErrNotFound) {
		_ = render.Render(w, r, apierror.ErrNotFound)
		return
	} else if errors.Is(err, task.ErrTaskState) {
		_ = render.Render(w, r, apierror.InvalidJobState("The job is already running."))
		return
	} else if err != nil {
		h.logger.Error("Could not start job.", "name", name, tint.Err(err))
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
