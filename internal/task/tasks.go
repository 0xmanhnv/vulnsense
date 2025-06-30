// Package task defines the structure and types for background tasks.
package task

import (
	"context"
	"log/slog"
	"os"
	"vulnsense/internal/usecase"

	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	// TypeFetchVulnerabilities is the task type for fetching vulnerabilities from all feeds.
	TypeFetchVulnerabilities = "task:fetch:vulnerabilities"
)

// NewFetchVulnerabilitiesTask creates a new task to fetch vulnerabilities.
// This task has no payload for now, as it triggers a system-wide fetch.
func NewFetchVulnerabilitiesTask() (*asynq.Task, error) {
	// For tasks without a payload, we can pass a nil payload.
	return asynq.NewTask(TypeFetchVulnerabilities, nil), nil
}

//---------------------------------------------------------------
// Task Handler: The logic that executes when a worker picks up a task.
//---------------------------------------------------------------

// HandleFetchVulnerabilitiesTask is the handler for TypeFetchVulnerabilities.
// It receives a use case from the worker's context and executes it.
func HandleFetchVulnerabilitiesTask(ctx context.Context, t *asynq.Task, fetchVulnsUseCase *usecase.FetchVulnerabilitiesUseCase) error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("task_type", t.Type())
	log.Info("Worker processing task")

	err := fetchVulnsUseCase.Execute(ctx)
	if err != nil {
		log.Error("Failed to execute fetch vulnerabilities use case", "error", err)
	}

	log.Info("Task processing completed")
	return err
}
