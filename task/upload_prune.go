package task

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

const TypePruneUploads = "upload:prune"

func NewPruneUploadsTask() *asynq.Task {
	return asynq.NewTask(TypePruneUploads, nil, asynq.TaskID(TypePruneUploads))
}

func (h *Handler) HandlePruneUploadsTask(_ context.Context, _ *asynq.Task) error {
	time.Sleep(2 * time.Minute)
	return nil
}
