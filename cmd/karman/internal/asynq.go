package internal

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
)

// AsynqLogLevel converts level into an equivalent asynq.LogLevel.
func AsynqLogLevel(level slog.Level) asynq.LogLevel {
	if level >= slog.LevelWarn {
		return asynq.ErrorLevel
	} else if level >= slog.LevelInfo {
		return asynq.WarnLevel
	} else if level >= slog.LevelDebug {
		return asynq.InfoLevel
	}
	return asynq.DebugLevel
}

// AsynqLogger is an implementation of asynq.Logger backed by a slog.Logger.
// This logger implements the Fatal method by sending a termination signal.
type AsynqLogger struct {
	Logger *slog.Logger     // underlying logger
	Sig    chan<- os.Signal // termination signal
}

// Debug logs a message at Debug level.
func (l *AsynqLogger) Debug(args ...any) {
	for _, line := range args {
		l.Logger.Debug(fmt.Sprintf("%s", line))
	}
}

// Info logs a message at Info level.
func (l *AsynqLogger) Info(args ...any) {
	for _, line := range args {
		l.Logger.Info(fmt.Sprintf("%s", line))
	}
}

// Warn logs a message at Warning level.
func (l *AsynqLogger) Warn(args ...any) {
	for _, line := range args {
		l.Logger.Warn(fmt.Sprintf("%s", line))
	}
}

// Error logs a message at Error level.
func (l *AsynqLogger) Error(args ...any) {
	for _, line := range args {
		l.Logger.Error(fmt.Sprintf("%s", line))
	}
}

// Fatal logs a message at Error level.
// Additionally Fatal sends a SIGINT to the current process, leading to the program's termination.
func (l *AsynqLogger) Fatal(args ...any) {
	for _, line := range args {
		l.Logger.Error(fmt.Sprintf("%s", line))
	}
	l.Sig <- os.Interrupt
}
