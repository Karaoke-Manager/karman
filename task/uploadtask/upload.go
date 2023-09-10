package uploadtask

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/Karaoke-Manager/karman/core/upload"
)

const (
	Queue            = "upload"
	TypePruneUploads = "upload:prune"
)

type Handler struct {
	mux *asynq.ServeMux

	logger *slog.Logger
	repo   upload.Repository
	store  upload.Store
}

func NewHandler(
	logger *slog.Logger,
	repo upload.Repository,
	store upload.Store,
) (string, *Handler) {
	mux := asynq.NewServeMux()
	h := &Handler{
		mux,
		logger,
		repo,
		store,
	}
	mux.HandleFunc(TypePruneUploads, h.ProcessPruneMediaTask)
	return "upload:", h
}

func (h *Handler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	return h.mux.ProcessTask(ctx, task)
}

func NewPruneUploadsTask() *asynq.Task {
	return asynq.NewTask(TypePruneUploads, nil, asynq.Queue(Queue), asynq.TaskID(TypePruneUploads))
}

func (h *Handler) ProcessPruneMediaTask(_ context.Context, _ *asynq.Task) error {
	time.Sleep(2 * time.Minute)
	return nil
}
