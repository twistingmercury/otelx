# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-12-17

### Added

- Core `otelx` library providing unified OpenTelemetry initialization with a
  single `Initialize()` function call
- `Telemetry` struct containing configured Logger, MeterProvider, and
  TracerProvider components
- Functional options pattern for flexible configuration:
  - `WithService()` for required service identity (name, version, environment)
  - `WithLogLevel()` and `WithLogWriter()` for logging configuration
  - `WithoutLogging()` to disable logging entirely
  - `WithMetrics()` and `WithMetricsPath()` for Prometheus metrics
    configuration
  - `WithTracing()`, `WithTraceSampleRate()`, and `WithTraceExporter()` for
    tracing configuration
  - `WithOTLPEndpoint()` and `WithOTLPInsecure()` for OTLP collector
    configuration
  - `WithAllSignals()` convenience option to enable metrics and tracing
    together
  - `WithDevelopmentDefaults()` preset for local development
  - `WithProductionDefaults()` preset for production deployments
- Structured logging via zerolog with automatic trace correlation through
  `LoggerWithSpan()` helper
- Prometheus metrics exposition via HTTP server with configurable port and path
- Distributed tracing via OTLP gRPC exporter with configurable sampling rates
- OpenTelemetry resource creation with standard semantic conventions
  (service.name, service.version, deployment.environment, host.name,
  process.pid, telemetry.sdk.*)
- Global OpenTelemetry provider registration for MeterProvider and
  TracerProvider
- W3C Trace Context and Baggage propagation for distributed tracing
- Graceful shutdown handling via `Telemetry.Shutdown()` to flush pending
  telemetry
- Convenience methods `Telemetry.Tracer()` and `Telemetry.Meter()` with no-op
  fallbacks
- Configuration validation with descriptive error messages
- Gin web framework middleware package (`middleware/gin`):
  - `LoggingMiddleware()` for trace-correlated request logging with
    status-based log levels
  - `CorrelationMiddleware()` for storing trace-correlated logger without
    request logging
  - `Logger()` and `MustLogger()` context helpers for retrieving loggers in
    handlers
  - Middleware options: `WithSkipPaths()`, `WithLogLevel()`,
    `WithRequestHeaders()`, `WithCustomFields()`
- Two example projects demonstrating otelx integration with Gin:
  - `gingonic-simple`: Embedded telemetry pattern for quick setup
  - `gingonic-decorator`: Decorator pattern for separation of concerns
- Docker Compose configurations for full observability stack (Prometheus,
  Jaeger, OpenTelemetry Collector)
- Comprehensive test coverage for core library and middleware

[Unreleased]: https://github.com/twistingmercury/otelx/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/twistingmercury/otelx/releases/tag/v1.0.0
