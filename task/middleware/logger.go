package middleware

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/lmittmann/tint"
)

// Logger is a task middleware that prints a log line when a task is started and when a task finishes.
func Logger(logger *slog.Logger) func(next asynq.Handler) asynq.Handler {
	return func(next asynq.Handler) asynq.Handler {
		fn := func(ctx context.Context, task *asynq.Task) error {
			id, _ := asynq.GetTaskID(ctx)
			queue, _ := asynq.GetQueueName(ctx)
			retry, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			logger.InfoContext(ctx, "Starting task.", "task", task.Type(), "taskID", id, "queue", queue, "retry", retry, "maxRetry", maxRetry)
			err := next.ProcessTask(ctx, task)
			if err != nil {
				logger.WarnContext(ctx, "Task did not complete successfully.", "task", task.Type(), "taskID", id, "queue", queue, "retry", retry, "maxRetry", maxRetry, tint.Err(err))
			} else {
				logger.InfoContext(ctx, "Task completed successfully.", "task", task.Type(), "taskID", id, "queue", queue, "retry", retry, "maxRetry", maxRetry)
			}
			return err
		}
		return asynq.HandlerFunc(fn)
	}
}
