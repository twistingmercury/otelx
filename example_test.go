package otelx_test

import (
	"context"
	"fmt"
	"os"

	"github.com/twistingmercury/otelx"
)

func Example() {
	ctx := context.Background()

	// Initialize with service identity, metrics, and tracing
	tel, err := otelx.Initialize(ctx,
		otelx.WithService("my-service", "1.0.0", "production"),
		otelx.WithMetrics(9090),
		otelx.WithTracing(),
		otelx.WithOTLPEndpoint("localhost:4317"),
		otelx.WithOTLPInsecure(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize telemetry: %v\n", err)
		return
	}
	defer func() { _ = tel.Shutdown(ctx) }()

	// Use the logger
	tel.Logger.Info().Msg("service started")
}

func Example_loggingOnly() {
	ctx := context.Background()

	// Initialize with just logging (no metrics or tracing)
	tel, err := otelx.Initialize(ctx,
		otelx.WithService("simple-service", "1.0.0", "development"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize telemetry: %v\n", err)
		return
	}
	defer func() { _ = tel.Shutdown(ctx) }()

	// Use structured logging
	tel.Logger.Info().
		Str("operation", "startup").
		Int("port", 8080).
		Msg("server starting")
}

func Example_metricsOnly() {
	ctx := context.Background()

	// Initialize with metrics only
	tel, err := otelx.Initialize(ctx,
		otelx.WithService("metrics-service", "1.0.0", "production"),
		otelx.WithMetrics(9090),
		otelx.WithMetricsPath("/prometheus"),
		otelx.WithoutLogging(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize telemetry: %v\n", err)
		return
	}
	defer func() { _ = tel.Shutdown(ctx) }()

	// Create metrics using the meter provider
	if tel.MeterProvider != nil {
		meter := tel.MeterProvider.Meter("my-service")
		counter, _ := meter.Int64Counter("http_requests_total")
		counter.Add(ctx, 1)
	}
}

func Example_development() {
	ctx := context.Background()

	// Development setup with debug logging and insecure connections
	tel, err := otelx.Initialize(ctx,
		otelx.WithService("dev-service", "0.0.1", "development"),
		otelx.WithDevelopmentDefaults(),
		otelx.WithMetrics(9090),
		otelx.WithTracing(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize telemetry: %v\n", err)
		return
	}
	defer tel.Shutdown(ctx)

	tel.Logger.Debug().Msg("debug logging is enabled")
}

func Example_production() {
	ctx := context.Background()

	// Production setup with lower sample rate and TLS
	tel, err := otelx.Initialize(ctx,
		otelx.WithService("prod-service", "2.0.0", "production"),
		otelx.WithProductionDefaults(),
		otelx.WithMetrics(9090),
		otelx.WithTracing(),
		otelx.WithOTLPEndpoint("otel-collector.internal:4317"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize telemetry: %v\n", err)
		return
	}
	defer tel.Shutdown(ctx)

	tel.Logger.Info().Msg("production service started")
}

func Example_customWriter() {
	ctx := context.Background()

	// Use a custom log writer (e.g., for testing or custom output)
	tel, err := otelx.Initialize(ctx,
		otelx.WithService("custom-service", "1.0.0", "test"),
		otelx.WithLogWriter(os.Stderr), // Write to stderr instead of stdout
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize telemetry: %v\n", err)
		return
	}
	defer tel.Shutdown(ctx)

	tel.Logger.Info().Msg("logging to stderr")
}

func ExampleWithAllSignals() {
	ctx := context.Background()

	// Enable all signals with default ports
	tel, err := otelx.Initialize(ctx,
		otelx.WithService("full-service", "1.0.0", "staging"),
		otelx.WithAllSignals(),
		otelx.WithOTLPEndpoint("localhost:4317"),
		otelx.WithOTLPInsecure(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize telemetry: %v\n", err)
		return
	}
	defer tel.Shutdown(ctx)

	// All telemetry components are now available
	tel.Logger.Info().Msg("all signals enabled")
}
