package nolog

import (
	"context"
	"log/slog"
)

// Handler is a noop logging handler.
var Handler slog.Handler = &handler{}

// Logger is a noop logger.
var Logger = slog.New(Handler)

// handler implements a noop slog.Handler.
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
