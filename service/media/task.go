package media

import (
	"context"

	"github.com/hibiken/asynq"
)

const (
	TaskQueue     = "media"
	TaskTypePrune = "media:prune"
)

type taskProvider struct{}

func NewPeriodicTaskConfigProvider() asynq.PeriodicTaskConfigProvider {
	return &taskProvider{}
}

func (*taskProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	return nil, nil
}

type taskHandler struct {
	repo  Repository
	store Store
}

func NewTaskHandler(repo Repository, store Store) (string, asynq.Handler) {
	h := &taskHandler{repo, store}
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskTypePrune, h.handlePruneTask)
	return "media:", mux
}

func NewPruneTask() *asynq.Task {
	panic("not implemented")
}

func (h *taskHandler) handlePruneTask(_ context.Context, _ *asynq.Task) error {
	panic("not implemented")
}
