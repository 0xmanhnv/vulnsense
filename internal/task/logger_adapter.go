package task

import "log/slog"

// SlogAsynqLogger is an adapter to make slog.Logger compatible with asynq.Logger.
type SlogAsynqLogger struct {
	logger *slog.Logger
}

func NewSlogAsynqLogger(logger *slog.Logger) *SlogAsynqLogger {
	return &SlogAsynqLogger{logger: logger}
}

func (l *SlogAsynqLogger) Debug(args ...any) {
	l.logger.Debug("asynq debug", "details", args)
}

func (l *SlogAsynqLogger) Info(args ...any) {
	l.logger.Info("asynq info", "details", args)
}

func (l *SlogAsynqLogger) Warn(args ...any) {
	l.logger.Warn("asynq warn", "details", args)
}

func (l *SlogAsynqLogger) Error(args ...any) {
	l.logger.Error("asynq error", "details", args)
}

func (l *SlogAsynqLogger) Fatal(args ...any) {
	l.logger.Error("asynq fatal", "details", args)
	// Asynq's Fatal is not expected to exit the process, so we just log as error.
}
