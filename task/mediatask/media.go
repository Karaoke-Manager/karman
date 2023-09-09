package mediatask

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/Karaoke-Manager/karman/service/media"
)

const (
	Queue          = "media"
	TypePruneMedia = "media:prune"
)

type Handler struct {
	mux *asynq.ServeMux

	logger *slog.Logger
	repo   media.Repository
	store  media.Store
}

func NewHandler(logger *slog.Logger, repo media.Repository, store media.Store) (string, *Handler) {
	mux := asynq.NewServeMux()
	h := &Handler{
		mux,
		logger,
		repo,
		store,
	}
	mux.HandleFunc(TypePruneMedia, h.ProcessPruneMediaTask)
	return "media:", h
}

func (h *Handler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	return h.mux.ProcessTask(ctx, task)
}

func NewPruneMediaTask() *asynq.Task {
	return asynq.NewTask(TypePruneMedia, nil, asynq.Queue(Queue), asynq.TaskID(TypePruneMedia))
}

func (h *Handler) ProcessPruneMediaTask(_ context.Context, _ *asynq.Task) error {
	time.Sleep(2 * time.Minute)
	return nil
}
