package observability

import (
	"context"

	"github.com/getsentry/sentry-go"
)

type Fields map[string]string

func Info(ctx context.Context, message string, fields Fields) {
	emit(ctx, "info", message, fields)
}

func Warn(ctx context.Context, message string, fields Fields) {
	emit(ctx, "warn", message, fields)
}

func Error(ctx context.Context, message string, err error, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	emit(ctx, "error", message, fields)
}

func emit(ctx context.Context, level string, message string, fields Fields) {
	if ctx == nil {
		ctx = context.Background()
	}

	logger := sentry.NewLogger(ctx)

	var entry sentry.LogEntry

	switch level {
	case "warn":
		entry = logger.Warn()
	case "error":
		entry = logger.Error()
	default:
		entry = logger.Info()
	}

	entry = entry.
		String("service", "disaster-alert-backend").
		String("log_type", "application")

	for key, value := range fields {
		if key == "" || value == "" {
			continue
		}

		entry = entry.String(key, value)
	}

	entry.Emit(message)
}
