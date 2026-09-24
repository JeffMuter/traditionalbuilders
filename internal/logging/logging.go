// Package logging carries a request-scoped *slog.Logger and request ID through
// context so every layer (handlers, store) can attach the same correlation ID
// to its log entries.
package logging

import (
	"context"
	"log/slog"
)

type contextKey int

const (
	loggerKey contextKey = iota
	requestIDKey
)

// WithLogger returns a context carrying l. Use FromContext to retrieve it.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext returns the request-scoped logger, or slog.Default() when none
// was installed (e.g. background jobs, tests).
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

// WithRequestID returns a context carrying id.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the request ID stored in ctx, or "" when absent.
func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
