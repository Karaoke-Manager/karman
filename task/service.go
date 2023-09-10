package task

import (
	"github.com/hibiken/asynq"

	"github.com/Karaoke-Manager/karman/task/mediatask"
)

type CronConfig struct {
	PruneMedia struct {
		Enabled  bool
		Schedule string
	}
}

type cronService struct {
	config    CronConfig
	inspector *asynq.Inspector
}

func NewCronService(config CronConfig, inspector *asynq.Inspector) (CronService, error) {
	return &cronService{config, inspector}, nil
}

func (s *cronService) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	return []*asynq.PeriodicTaskConfig{{
		Cronspec: "* * * * *",
		Task:     mediatask.NewPruneMediaTask(),
	}}, nil
}

func (s *cronService) ListJobs() (map[string]JobStat, error) {
	// The service should probably know about all scheduled tasks and possibly all tasks.
	// This probably returns a static list
	// Each task may be disabled by the server config
	//TODO implement me
	panic("implement me")
}

func (s *cronService) RunJob(id string) error {
	//TODO implement me
	panic("implement me")
}

func (s *cronService) StopJob(id string) error {
	//TODO implement me
	panic("implement me")
}

func (s *cronService) StatJob(id string) (JobStat, error) {
	info, err := s.inspector.GetTaskInfo("", id)
	//TODO implement me
	panic("implement me")
}
