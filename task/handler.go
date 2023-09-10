package task

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/lmittmann/tint"

	"github.com/Karaoke-Manager/karman/core/media"
	"github.com/Karaoke-Manager/karman/task/mediatask"
)

// Handler is the main asynq.Handler.
// It consists of the sub-handlers for various task areas.
type Handler struct {
	logger *slog.Logger
	mux    *asynq.ServeMux
}

// NewHandler creates a new Handler instance that can process tasks.
func NewHandler(
	logger *slog.Logger,
	repo media.Repository,
	store media.Store,
) *Handler {
	mux := asynq.NewServeMux()
	h := &Handler{
		logger,
		mux,
	}
	mux.Use(h.Logger)
	mux.Handle(mediatask.NewHandler(
		logger,
		repo,
		store,
	))
	return h
}

// ProcessTask begins processing the specified task or returns an error.
func (h *Handler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	return h.mux.ProcessTask(ctx, task)
}

// Logger is a task middleware that prints a log line when a task is started and when a task finishes.
func (h *Handler) Logger(next asynq.Handler) asynq.Handler {
	fn := func(ctx context.Context, task *asynq.Task) error {
		id, _ := asynq.GetTaskID(ctx)
		queue, _ := asynq.GetQueueName(ctx)
		retry, _ := asynq.GetRetryCount(ctx)
		maxRetry, _ := asynq.GetMaxRetry(ctx)
		h.logger.InfoContext(ctx, "Starting task.", "task", task.Type(), "taskID", id, "queue", queue, "retry", retry, "maxRetry", maxRetry)
		err := next.ProcessTask(ctx, task)
		if err != nil {
			h.logger.WarnContext(ctx, "Task did not complete successfully.", "task", task.Type(), "taskID", id, "queue", queue, "retry", retry, "maxRetry", maxRetry, tint.Err(err))
		} else {
			h.logger.InfoContext(ctx, "Task completed successfully.", "task", task.Type(), "taskID", id, "queue", queue, "retry", retry, "maxRetry", maxRetry)
		}
		return err
	}
	return asynq.HandlerFunc(fn)
}
