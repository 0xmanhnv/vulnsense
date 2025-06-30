// Package task handles the setup and execution of background jobs.
package task

import (
	"context"
	"log/slog"

	"vulnsense/internal/app"

	"github.com/hibiken/asynq"
)

const (
	// This is a queue name. You can have multiple queues with different priorities.
	// For now, we'll use one "critical" queue.
	QueueCritical = "critical"
)

// Processor runs the Asynq worker server.
type Processor struct {
	server *asynq.Server
	logger *slog.Logger
}

// NewProcessor creates a new task processor.
func NewProcessor(redisOpt asynq.RedisClientOpt, app *app.App, logger *slog.Logger) *Processor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 6, // 6 is the highest priority
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error("Asynq task failed", "task_type", task.Type(), "payload", string(task.Payload()), "error", err)
			}),
			Logger: NewSlogAsynqLogger(logger),
		},
	)

	return &Processor{
		server: server,
		logger: logger,
	}
}

// Start registers task handlers and starts the worker server.
func (p *Processor) Start(app *app.App) error {
	mux := asynq.NewServeMux()

	// Register handlers here.
	// The handler function is wrapped to inject dependencies from the app struct.
	mux.HandleFunc(
		TypeFetchVulnerabilities,
		func(ctx context.Context, t *asynq.Task) error {
			return HandleFetchVulnerabilitiesTask(ctx, t, app.FetchVulnsUseCase())
		},
	)
	// ... register other task handlers here

	p.logger.Info("Starting Asynq worker server")
	return p.server.Run(mux)
}
