package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/lmittmann/tint"
)

const TypeProcessUpload = "upload:process"

func NewProcessUploadTask(id uuid.UUID) *asynq.Task {
	return asynq.NewTask(TypeProcessUpload, id[:], asynq.TaskID(fmt.Sprintf("%s:%s", TypeProcessUpload, id)))
}

func (h *Handler) HandleProcessUploadTask(ctx context.Context, task *asynq.Task) error {
	id, err := uuid.FromBytes(task.Payload())
	if err != nil {
		h.logger.WarnContext(ctx, "Could not process upload.", "uuid", id, tint.Err(err))
		return errors.Join(err, ErrInvalidPayload)
	}
	return h.uploadService.ProcessUpload(ctx, id)
}
