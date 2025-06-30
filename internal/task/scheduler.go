// Package task handles the setup and execution of background jobs.
package task

import (
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
)

// Scheduler runs the Asynq scheduler.
type Scheduler struct {
	scheduler *asynq.Scheduler
	logger    *slog.Logger
}

// NewScheduler creates a new task scheduler.
func NewScheduler(redisOpt asynq.RedisClientOpt, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		scheduler: asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
			Location: time.Local,
			Logger:   NewSlogAsynqLogger(logger),
		}),
		logger: logger,
	}
}

// Start registers periodic tasks and starts the scheduler.
func (s *Scheduler) Start() error {
	// Register tasks here.

	// Task to fetch vulnerabilities every 24 hours.
	task, err := NewFetchVulnerabilitiesTask()
	if err != nil {
		return err
	}
	// The cron spec "@daily" runs once a day at midnight.
	entryID, err := s.scheduler.Register("@daily", task)
	if err != nil {
		return err
	}
	s.logger.Info("Registered periodic task", "entry_id", entryID, "task_type", task.Type(), "cron_spec", "@daily")

	// ... register other periodic tasks here

	s.logger.Info("Starting Asynq scheduler")
	return s.scheduler.Run()
}
