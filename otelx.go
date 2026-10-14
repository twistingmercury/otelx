package otelx

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Telemetry holds the components created by Initialize.
type Telemetry struct {
	// Logger is a zerolog logger tagged with the service identity. Use
	// LoggerWithSpan to add trace and span IDs to it.
	Logger zerolog.Logger

	// MeterProvider is nil when metrics are disabled.
	MeterProvider *sdkmetric.MeterProvider

	// TracerProvider is nil when tracing is disabled.
	TracerProvider *sdktrace.TracerProvider

	metrics *metricsServer
}

// Initialize creates the logger, and the meter and tracer providers that cfg
// enables, and registers the providers as the OpenTelemetry globals.
// Call Shutdown on the result to flush and release them.
//
// Example:
//
//	tel, err := otelx.Initialize(ctx, otelx.Config{
//	    ServiceName:    "my-service",
//	    ServiceVersion: "1.0.0",
//	    Environment:    "production",
//	    MetricsAddr:    ":9090",
//	    Tracing:        true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer tel.Shutdown(ctx)
func Initialize(ctx context.Context, cfg Config) (*Telemetry, error) {
	cfg, err := cfg.withDefaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp, err := setupTracing(ctx, cfg, res)
	if err != nil {
		return nil, fmt.Errorf("failed to setup tracing: %w", err)
	}

	mp, metrics, err := setupMetrics(cfg, res)
	if err != nil {
		if tp != nil {
			_ = tp.Shutdown(ctx)
		}
		return nil, fmt.Errorf("failed to setup metrics: %w", err)
	}

	if tp != nil {
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagator)
	}
	if mp != nil {
		otel.SetMeterProvider(mp)
	}

	return &Telemetry{
		Logger:         newLogger(cfg, res),
		MeterProvider:  mp,
		TracerProvider: tp,
		metrics:        metrics,
	}, nil
}

// Shutdown flushes and stops the tracer provider, the meter provider and the
// metrics server. It returns every failure, joined with errors.Join.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error

	if t.TracerProvider != nil {
		if err := t.TracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer provider shutdown: %w", err))
		}
	}
	if t.MeterProvider != nil {
		if err := t.MeterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
		}
	}
	if t.metrics != nil {
		if err := t.metrics.server.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metrics server shutdown: %w", err))
		}
	}

	return errors.Join(errs...)
}

// MetricsURL returns the URL metrics are served on, or "" when metrics are
// disabled. The host is the one the server listens on.
func (t *Telemetry) MetricsURL() string {
	if t.metrics == nil {
		return ""
	}
	return t.metrics.url
}
