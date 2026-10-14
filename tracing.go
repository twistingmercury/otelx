package otelx

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// propagator is registered globally when tracing is enabled.
var propagator = propagation.NewCompositeTextMapPropagator(
	propagation.TraceContext{},
	propagation.Baggage{},
)

// setupTracing returns a TracerProvider, or nil when tracing is disabled.
// The OTLP exporter reads its endpoint, headers and TLS settings from the
// OTEL_EXPORTER_OTLP_* environment variables, and the provider reads its
// sampler from OTEL_TRACES_SAMPLER unless cfg.Sampler is set.
func setupTracing(ctx context.Context, cfg Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	if !cfg.Tracing {
		return nil, nil
	}

	exporter := cfg.TraceExporter
	if exporter == nil {
		var err error
		exporter, err = otlptracegrpc.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(cfg.Sampler), // nil keeps the environment's sampler
	), nil
}
