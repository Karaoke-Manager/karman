package internal

import (
	"context"
	"fmt"
	"log/slog"
)

type AsynqLogger struct {
	l      *slog.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

func NewAsynqLogger(l *slog.Logger, component string) *AsynqLogger {
	ctx, cancel := context.WithCancel(context.Background())
	return &AsynqLogger{
		l:      l.With("component", component),
		ctx:    ctx,
		cancel: cancel,
	}
}

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

// Fatal logs a message at Fatal level
// and process will exit with status set to 1.
func (l *AsynqLogger) Fatal(args ...any) {
	for _, line := range args {
		l.l.Error(fmt.Sprintf("%s", line))
	}
	l.cancel()
}
