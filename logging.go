package otelx

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// setupLogging configures the zerolog logger.
// If tracing is enabled, the logger can include trace context via the
// WithTraceContext helper.
func setupLogging(cfg *Config) zerolog.Logger {
	if !cfg.LoggingEnabled {
		return zerolog.Nop()
	}

	// Configure zerolog output
	output := cfg.LogWriter
	if output == nil {
		// This shouldn't happen if validation passed, but be defensive
		return zerolog.Nop()
	}

	// Set global zerolog level
	zerolog.SetGlobalLevel(cfg.LogLevel)

	// Create base logger with service context
	logger := zerolog.New(output).With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("version", cfg.ServiceVersion).
		Str("environment", cfg.Environment).
		Logger()

	// Add caller info for debug level
	if cfg.LogLevel <= zerolog.DebugLevel {
		logger = logger.With().Caller().Logger()
	}

	return logger
}

// LoggerWithSpan creates a new logger context that includes trace and span IDs
// from the given context's span. This enables log correlation with traces.
func LoggerWithSpan(logger zerolog.Logger, ctx context.Context) zerolog.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return logger
	}

	return logger.With().
		Str("trace_id", span.SpanContext().TraceID().String()).
		Str("span_id", span.SpanContext().SpanID().String()).
		Logger()
}

// LogLevel converts a string log level to zerolog.Level.
// Supported values: "trace", "debug", "info", "warn", "error", "fatal", "panic", "disabled".
// Returns zerolog.InfoLevel for unrecognized values.
func LogLevel(level string) zerolog.Level {
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled", "none", "off":
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}
