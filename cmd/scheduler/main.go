// Command scheduler is a standalone application that enqueues periodic tasks.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"vulnsense/internal/config"
	"vulnsense/internal/task"

	"github.com/hibiken/asynq"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Starting scheduler process")

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	redisConnectionOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// Create a cancellable context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	scheduler := task.NewScheduler(redisConnectionOpt, logger)
	if err := scheduler.Start(); err != nil {
		logger.Error("failed to start scheduler", "error", err)
		os.Exit(1)
	}

	<-ctx.Done() // Wait for shutdown signal
	logger.Info("Scheduler process shutting down")
}
