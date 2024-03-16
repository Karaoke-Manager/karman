// Package task
// TODO: Doc: Common errors
package task

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/Karaoke-Manager/karman/core/media"
	"github.com/Karaoke-Manager/karman/core/upload"
	"github.com/Karaoke-Manager/karman/task/middleware"
)

var (
	// ErrInvalidPayload indicates that the payload of a task did not correspond to the expected schema.
	// Tasks with this error will not be retried.
	ErrInvalidPayload = fmt.Errorf("invalid payload: %w", asynq.SkipRetry)
)

// Handler implements the asynq.Handler interface.
type Handler struct {
	logger *slog.Logger
	mux    *asynq.ServeMux

	mediaRepo     media.Repository
	mediaService  media.Service
	uploadService upload.Service
	uploadRepo    upload.Repository
	uploadStore   upload.Store
}

// NewHandler creates a new Handler instance that can process tasks.
func NewHandler(
	logger *slog.Logger,
	mediaRepo media.Repository,
	mediaService media.Service,
	uploadService upload.Service,
	uploadRepo upload.Repository,
	uploadStore upload.Store,
) *Handler {
	mux := asynq.NewServeMux()
	h := &Handler{
		logger,
		mux,
		mediaRepo,
		mediaService,
		uploadService,
		uploadRepo,
		uploadStore,
	}
	mux.Use(middleware.Logger(h.logger))
	mux.HandleFunc(TypePruneMedia, h.HandlePruneMediaTask)
	mux.HandleFunc(TypePruneUploads, h.HandlePruneUploadsTask)
	mux.HandleFunc(TypeProcessUpload, h.HandleProcessUploadTask)
	return h
}

// ProcessTask begins processing the specified task or returns an error.
func (h *Handler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	return h.mux.ProcessTask(ctx, task)
}
