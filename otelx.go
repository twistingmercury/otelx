package otelx

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Telemetry holds initialized OpenTelemetry components.
// It provides access to the logger, meter provider, and tracer provider
// for creating metrics, traces, and logs in your application.
type Telemetry struct {
	// Logger is the configured zerolog logger with trace correlation.
	Logger zerolog.Logger

	// MeterProvider is the OpenTelemetry MeterProvider for creating metrics.
	// Will be nil if metrics are not enabled.
	MeterProvider *sdkmetric.MeterProvider

	// TracerProvider is the OpenTelemetry TracerProvider for creating traces.
	// Will be nil if tracing is not enabled.
	TracerProvider *sdktrace.TracerProvider

	// metricsServer holds the HTTP server serving Prometheus metrics.
	metricsServer *MetricsServer

	// config holds the configuration used for initialization.
	config *Config
}

// Initialize creates and configures all telemetry components based on the provided options.
// It returns a Telemetry struct containing the initialized components, or an error if
// initialization fails.
//
// The WithService option is required and must be provided to identify the service.
//
// Example:
//
//	tel, err := otelx.Initialize(ctx,
//	    otelx.WithService("my-service", "1.0.0", "production"),
//	    otelx.WithMetrics(9090),
//	    otelx.WithTracing(),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer tel.Shutdown(ctx)
func Initialize(ctx context.Context, opts ...Option) (*Telemetry, error) {
	// Start with default configuration
	cfg := NewDefaultConfig()

	// Apply all options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create OpenTelemetry resource
	res, err := NewResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize tracing first (needed for log correlation)
	tp, err := setupTracing(ctx, cfg, res)
	if err != nil {
		return nil, fmt.Errorf("failed to setup tracing: %w", err)
	}

	// Initialize logging
	logger := setupLogging(cfg)

	// Initialize metrics
	mp, metricsServer, err := setupMetrics(ctx, cfg, res)
	if err != nil {
		// Clean up tracing if metrics setup fails
		if tp != nil {
			_ = tp.Shutdown(ctx)
		}
		return nil, fmt.Errorf("failed to setup metrics: %w", err)
	}

	// Set global meter provider if enabled
	if mp != nil {
		otel.SetMeterProvider(mp)
	}

	return &Telemetry{
		Logger:         logger,
		MeterProvider:  mp,
		TracerProvider: tp,
		metricsServer:  metricsServer,
		config:         cfg,
	}, nil
}

// Shutdown gracefully shuts down all telemetry components.
// It should be called when the application is shutting down to ensure
// all pending telemetry data is flushed and resources are released.
//
// Example:
//
//	defer tel.Shutdown(ctx)
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error

	// Shutdown tracing first to flush any pending spans
	if t.TracerProvider != nil {
		if err := t.TracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer provider shutdown: %w", err))
		}
	}

	// Shutdown metrics
	if t.MeterProvider != nil {
		if err := t.MeterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
		}
	}

	// Shutdown metrics HTTP server
	if t.metricsServer != nil {
		if err := t.metricsServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metrics server shutdown: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// MetricsURL returns the URL where metrics are being served.
// Returns an empty string if metrics are not enabled.
func (t *Telemetry) MetricsURL() string {
	if t.metricsServer == nil {
		return ""
	}
	return t.metricsServer.URL()
}

// Config returns a copy of the configuration used for initialization.
func (t *Telemetry) Config() Config {
	if t.config == nil {
		return Config{}
	}
	return *t.config.Clone()
}

// Tracer returns a tracer from the TracerProvider with the given name.
// If tracing is not enabled, it returns a no-op tracer from the global provider.
func (t *Telemetry) Tracer(name string) trace.Tracer {
	if t.TracerProvider == nil {
		return otel.Tracer(name)
	}
	return t.TracerProvider.Tracer(name)
}

// Meter returns a meter from the MeterProvider with the given name.
// If metrics are not enabled, it returns a no-op meter from the global provider.
func (t *Telemetry) Meter(name string) metric.Meter {
	if t.MeterProvider == nil {
		return otel.Meter(name)
	}
	return t.MeterProvider.Meter(name)
}
