package internal

import (
	"context"
	"fmt"
	"log/slog"

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
// This logger implements the Fatal method by cancelling a context.
type AsynqLogger struct {
	l      *slog.Logger
	ctx    context.Context
	cancel context.CancelCauseFunc
}

// NewAsynqLogger creates a new asynq.Logger instance backed by l.
// l will always add the "log" attribute with the specified value.
func NewAsynqLogger(l *slog.Logger, log string) *AsynqLogger {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &AsynqLogger{
		l:      l.With("log", log),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Context returns the context of l.
// If l.Fatal has been called, the context will be cancelled.
func (l *AsynqLogger) Context() context.Context {
	return l.ctx
}

// Debug logs a message at Debug level.
func (l *AsynqLogger) Debug(args ...any) {
	for _, line := range args {
		l.l.Debug(fmt.Sprintf("%s", line))
	}
}

// Info logs a message at Info level.
func (l *AsynqLogger) Info(args ...any) {
	for _, line := range args {
		l.l.Info(fmt.Sprintf("%s", line))
	}
}

// Warn logs a message at Warning level.
func (l *AsynqLogger) Warn(args ...any) {
	for _, line := range args {
		l.l.Warn(fmt.Sprintf("%s", line))
	}
}

// Error logs a message at Error level.
func (l *AsynqLogger) Error(args ...any) {
	for _, line := range args {
		l.l.Error(fmt.Sprintf("%s", line))
	}
}

// Fatal logs a message at Error level.
// Additionally Fatal cancels the context of l, leading to the program's termination.
func (l *AsynqLogger) Fatal(args ...any) {
	for _, line := range args {
		l.l.Error(fmt.Sprintf("%s", line))
	}
	var err error
	if len(args) > 0 {
		err = fmt.Errorf("%s", args[0])
	}
	l.cancel(err)
}
