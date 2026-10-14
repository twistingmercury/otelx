// Package otelx provides a unified initialization API for OpenTelemetry
// telemetry signals including logging, metrics, and tracing.
//
// otelx is the spiritual successor to github.com/twistingmercury/telemetry/v2,
// providing a simpler, more idiomatic Go API for initializing and managing
// OpenTelemetry components in your applications.
//
// # Quick Start
//
// The simplest way to get started is to use Initialize with the required
// service identity option:
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
//
// # Components
//
// otelx initializes three telemetry signals:
//
//   - Logging: Structured logging via zerolog with OpenTelemetry trace correlation
//   - Metrics: Prometheus metrics exposition via an HTTP server
//   - Tracing: Distributed tracing via OTLP gRPC exporter
//
// Each signal can be enabled or disabled independently using options.
//
// # Configuration Options
//
// otelx uses functional options for configuration:
//
//   - WithService: Required. Sets service name, version, and environment
//   - WithLogLevel: Sets the minimum log level (default: Info)
//   - WithLogWriter: Sets the log output writer (default: os.Stdout)
//   - WithoutLogging: Disables logging entirely
//   - WithMetrics: Enables Prometheus metrics on the specified port
//   - WithMetricsPath: Sets the metrics endpoint path (default: /metrics)
//   - WithTracing: Enables OTLP tracing
//   - WithTraceSampleRate: Sets the trace sampling rate (0.0-1.0)
//   - WithTraceExporter: Uses a custom trace exporter
//   - WithOTLPEndpoint: Sets the OTLP collector endpoint
//   - WithOTLPInsecure: Disables TLS for OTLP connections
//   - WithAllSignals: Enables metrics (port 9090) and tracing
//   - WithDevelopmentDefaults: Enables debug logging, pretty output, insecure connections
//   - WithProductionDefaults: Enables info logging, JSON output, TLS, 10% sampling
//
// # Resource Attributes
//
// otelx automatically creates an OpenTelemetry resource with standard
// semantic conventions including:
//
//   - service.name
//   - service.version
//   - deployment.environment.name
//   - host.name
//   - process.pid
//   - telemetry.sdk.name, telemetry.sdk.version, telemetry.sdk.language
//     (as reported by the OpenTelemetry Go SDK)
//
// Attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME are applied
// last and override the values passed to WithService.
//
// # Global State
//
// Initialize sets these process-wide globals:
//
//   - otel.SetTracerProvider, when tracing is enabled
//   - otel.SetTextMapPropagator (W3C Trace Context and Baggage), when tracing is enabled
//   - otel.SetMeterProvider, when metrics are enabled
//
// It does not change zerolog's global level or use the default Prometheus
// registry: each Telemetry has its own registry and metrics server. Calling
// Initialize more than once is safe, but the globals above point at the most
// recent call's providers.
//
// # Thread Safety
//
// The Telemetry struct and all its components are safe for concurrent use.
//
// # Shutdown
//
// Always call Shutdown on the returned Telemetry struct to ensure proper
// cleanup of resources:
//
//	defer tel.Shutdown(ctx)
//
// Shutdown will (returning every failure, joined with errors.Join):
//   - Flush and close the trace provider
//   - Flush and close the meter provider
//   - Stop the metrics HTTP server (if running)
package otelx
