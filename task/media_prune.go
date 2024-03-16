package task

import (
	"context"
	"encoding/binary"

	"github.com/hibiken/asynq"
	"github.com/lmittmann/tint"
)

// TypePruneMedia is the task type for the prune media task.
// This task detects and deletes media files that are not referenced by any songs.
// Deletion is permanent (i.e. not a soft-delete).
//
// The payload of the task is a single int64 in varint encoding specifying the maximum number of media records to be deleted.
//
// Only a single task of this type should be active at a time.
// Multiple active tasks will try to delete the same media records.
const TypePruneMedia = "media:prune"

// NewPruneMediaTask creates a new [TypePruneMedia] task.
// At most limit media records will be deleted when this task is executed.
func NewPruneMediaTask(limit int64) *asynq.Task {
	payload := binary.AppendVarint(nil, limit)
	return asynq.NewTask(TypePruneMedia, payload)
}

// HandlePruneMediaTask handles [TypePruneMedia] tasks.
func (h *Handler) HandlePruneMediaTask(ctx context.Context, task *asynq.Task) error {
	limit, n := binary.Varint(task.Payload())
	if n <= 0 || limit <= 0 {
		return ErrInvalidPayload
	}
	files, err := h.mediaRepo.FindOrphanedFiles(ctx, limit)
	if err != nil {
		h.logger.WarnContext(ctx, "Could not find prunable media files.", tint.Err(err))
		return err
	}
	for _, file := range files {
		if err = h.mediaService.DeleteFile(ctx, file.UUID); err != nil {
			h.logger.WarnContext(ctx, "Could not delete prunable media files.", tint.Err(err))
			return err
		}
	}
	return nil
}
