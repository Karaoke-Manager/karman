package task

import (
	"github.com/hibiken/asynq"

	"github.com/Karaoke-Manager/karman/task/mediatask"
)

const (
	QueueMedia  = mediatask.Queue
	QueueUpload = "upload"
)

type CronService interface {
	asynq.PeriodicTaskConfigProvider

	ListJobs()
	RunJob(id string)
	StopJob(id string)
	StatJob(id string) // return if it's running or when it is scheduled. Return basically a truncated asynq.TaskInfo.
}
