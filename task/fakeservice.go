package task

import (
	"time"

	"github.com/hibiken/asynq"

	"github.com/Karaoke-Manager/karman/core"
)

type fakeCronService struct {
	jobs []string
}

func NewFakeCronService(jobs ...string) CronService {
	return &fakeCronService{jobs}
}

func (s fakeCronService) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	return nil, nil
}

func (s fakeCronService) ListJobs() (map[string]JobStat, error) {
	stats := make(map[string]JobStat, len(s.jobs))
	for _, name := range s.jobs {
		stats[name] = JobStat{
			Enabled:     true,
			Active:      false,
			ScheduledAt: time.Now().Add(10 * time.Minute),
		}
	}
	return stats, nil
}

func (s fakeCronService) RunJob(id string) error {
	for _, name := range s.jobs {
		if name == id {
			return nil
		}
	}
	return core.ErrNotFound
}

func (s fakeCronService) StatJob(id string) (JobStat, error) {
	for _, name := range s.jobs {
		if name == id {
			return JobStat{
				Enabled:     true,
				Active:      false,
				ScheduledAt: time.Now().Add(10 * time.Minute),
			}, nil
		}
	}
	return JobStat{}, core.ErrNotFound
}
