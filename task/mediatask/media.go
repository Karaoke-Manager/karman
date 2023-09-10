package mediatask

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/lmittmann/tint"

	"github.com/Karaoke-Manager/karman/core/media"
)

const (
	Queue          = "media"
	TypePruneMedia = "media:prune"
)

type Handler struct {
	mux *asynq.ServeMux

	logger  *slog.Logger
	repo    media.Repository
	service media.Service
}

func NewHandler(
	logger *slog.Logger,
	repo media.Repository,
	service media.Service,
) (string, *Handler) {
	mux := asynq.NewServeMux()
	h := &Handler{
		mux,
		logger,
		repo,
		service,
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

func (h *Handler) ProcessPruneMediaTask(ctx context.Context, _ *asynq.Task) error {
	// 100 files at a time should be enough.
	// If for some reason there are more orphaned files, they will be deleted on the next run.
	// It seems very unlikely that there will be 100s of orphaned files continuously.
	files, err := h.repo.FindOrphanedFiles(ctx, 100)
	if err != nil {
		h.logger.WarnContext(ctx, "Could not prune media files.", tint.Err(err))
		return err
	}
	for _, file := range files {
		if err = h.service.DeleteFile(ctx, file.UUID); err != nil {
			return err
		}
	}
	return nil
}
