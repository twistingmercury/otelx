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
//   - deployment.environment
//   - host.name
//   - process.pid
//   - telemetry.sdk.name
//   - telemetry.sdk.version
//   - telemetry.sdk.language
//
// # Thread Safety
//
// The Telemetry struct and all its components are safe for concurrent use.
// The global OpenTelemetry providers are set during initialization and
// should not be modified afterward.
//
// # Shutdown
//
// Always call Shutdown on the returned Telemetry struct to ensure proper
// cleanup of resources:
//
//	defer tel.Shutdown(ctx)
//
// Shutdown will:
//   - Flush and close the trace provider
//   - Flush and close the meter provider
//   - Stop the metrics HTTP server (if running)
package otelx
