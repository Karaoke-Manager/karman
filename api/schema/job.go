package schema

import (
	"time"

	"github.com/Karaoke-Manager/karman/task"
)

type Job struct {
	Enabled     bool      `json:"enabled"`
	Active      bool      `json:"active"`
	ScheduledAt time.Time `json:"scheduledAt"`
}

func FromJobStat(stat task.JobStat) Job {
	return Job(stat)
}
