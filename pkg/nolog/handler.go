package nolog

import (
	"context"
	"log/slog"
)

var Handler slog.Handler = &handler{}
var Logger = slog.New(Handler)

type handler struct{}

// Enabled returns false for all log levels.
func (*handler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

// Handle is a noop that always returns nil.
func (*handler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

// WithAttrs returns h unmodified.
func (h *handler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns h unmodified.
func (h *handler) WithGroup(_ string) slog.Handler {
	return h
}
