package task

import (
	"github.com/hibiken/asynq"

	"github.com/Karaoke-Manager/karman/task/mediatask"
)

type cronService struct {
	inspector *asynq.Inspector
}

func NewCronService() CronService {
	return &cronService{}
}

func (s *cronService) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	return []*asynq.PeriodicTaskConfig{{
		Cronspec: "* * * * *",
		Task:     mediatask.NewPruneMediaTask(),
	}}, nil
}

func (s *cronService) ListJobs() {
	// The service should probably know about all scheduled tasks and possibly all tasks.
	// This probably returns a static list
	// Each task may be disabled by the server config
	//TODO implement me
	panic("implement me")
}

func (s *cronService) RunJob(id string) {
	//TODO implement me
	panic("implement me")
}

func (s *cronService) StopJob(id string) {
	//TODO implement me
	panic("implement me")
}

func (s *cronService) StatJob(id string) {
	//TODO implement me
	panic("implement me")
}
