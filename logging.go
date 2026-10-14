package otelx

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

// newLogger returns a JSON zerolog logger tagged with the service identity
// from res. The level applies to this logger only; zerolog's global level is
// left alone.
func newLogger(cfg Config, res *resource.Resource) zerolog.Logger {
	level, _ := zerolog.ParseLevel(cfg.LogLevel) // validated in withDefaults

	ctx := zerolog.New(cfg.LogWriter).Level(level).With().Timestamp()
	for _, f := range []struct {
		name string
		key  attribute.Key
	}{
		{"service", semconv.ServiceNameKey},
		{"version", semconv.ServiceVersionKey},
		{"environment", semconv.DeploymentEnvironmentNameKey},
	} {
		if v, ok := res.Set().Value(f.key); ok {
			ctx = ctx.Str(f.name, v.String())
		}
	}
	logger := ctx.Logger()

	if level <= zerolog.DebugLevel {
		logger = logger.With().Caller().Logger()
	}
	return logger
}

// LoggerWithSpan returns logger with trace_id and span_id fields taken from
// the span in ctx. It returns logger unchanged when ctx has no valid span.
func LoggerWithSpan(logger zerolog.Logger, ctx context.Context) zerolog.Logger {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return logger
	}

	return logger.With().
		Str("trace_id", sc.TraceID().String()).
		Str("span_id", sc.SpanID().String()).
		Logger()
}
