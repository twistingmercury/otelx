package otelx

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// setupTracing configures the OTLP trace exporter and returns a TracerProvider.
func setupTracing(ctx context.Context, cfg *Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	if !cfg.TracingEnabled {
		return nil, nil
	}

	var exporter sdktrace.SpanExporter
	var err error

	if cfg.TraceExporter != nil {
		exporter = cfg.TraceExporter
	} else {
		// Create OTLP gRPC exporter
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		}

		if cfg.OTLPInsecure {
			opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
			opts = append(opts, otlptracegrpc.WithInsecure())
		}

		exporter, err = otlptracegrpc.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}
	}

	// The sample rate applies to root spans only; child spans follow their
	// parent's sampling decision so distributed traces stay complete.
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceSampleRate))

	// Create TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for distributed tracing
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}
