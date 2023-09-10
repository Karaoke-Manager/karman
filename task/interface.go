package task

import (
	"errors"
	"time"

	"github.com/hibiken/asynq"

	"github.com/Karaoke-Manager/karman/task/mediatask"
	"github.com/Karaoke-Manager/karman/task/uploadtask"
)

const (
	QueueMedia  = mediatask.Queue
	QueueUpload = uploadtask.Queue
)

var (
	// ErrTaskState indicates that a task was not in a state where the action could be performed.
	ErrTaskState = errors.New("invalid task state")
)

type JobStat struct {
	Enabled     bool
	Active      bool
	ScheduledAt time.Time
}

type CronService interface {
	asynq.PeriodicTaskConfigProvider

	ListJobs() (map[string]JobStat, error)
	RunJob(id string) error
	StatJob(id string) (JobStat, error)
}
