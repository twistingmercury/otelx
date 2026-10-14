# Changelog

<!-- markdownlint-disable MD024 -->

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Breaking: v2

The module path is now `github.com/twistingmercury/otelx/v2`, and
`Initialize` takes a `Config` struct instead of functional options.

| v1 | v2 |
| --- | --- |
| `WithService(name, version, env)` | `Config.ServiceName`, `ServiceVersion`, `Environment` (only the name is required, and `OTEL_SERVICE_NAME` can supply it) |
| `WithLogLevel(zerolog.Level)` | `Config.LogLevel` as a level name, such as `"debug"` |
| `WithLogWriter(w)` | `Config.LogWriter` |
| `WithoutLogging()` | `Config.LogLevel: "disabled"` |
| `WithMetrics(port)` | `Config.MetricsAddr`, such as `":9090"` |
| `WithMetricsPath(path)` | `Config.MetricsPath` |
| `WithTracing()` | `Config.Tracing: true` |
| `WithTraceExporter(e)` | `Config.TraceExporter` |
| `WithTraceSampleRate(r)` | `OTEL_TRACES_SAMPLER=parentbased_traceidratio` with `OTEL_TRACES_SAMPLER_ARG=r`, or `Config.Sampler` |
| `WithOTLPEndpoint(e)` | `OTEL_EXPORTER_OTLP_ENDPOINT` |
| `WithOTLPInsecure()` | `OTEL_EXPORTER_OTLP_INSECURE=true`, or an `http://` endpoint |
| `WithAllSignals`, `WithDevelopmentDefaults`, `WithProductionDefaults` | Removed; set the fields directly |
| `tel.Tracer(name)`, `tel.Meter(name)` | `otel.Tracer(name)`, `otel.Meter(name)`, or the provider fields |
| `tel.Config()` | Removed |
| `NewDefaultConfig`, `Config.Validate`, `Config.Clone` | Removed |
| `NewResource`, `MetricsServer`, `LogLevel`, `Default*` constants | Unexported or removed |

- The gin middleware is now its own module,
  `github.com/twistingmercury/otelx/middleware/gin/v2`, so importing otelx no
  longer pulls in gin, sonic, quic-go or mongo-driver
- The gin request log no longer includes the raw query string, which could
  carry tokens or personal data
- The default OTLP exporter no longer overrides `OTEL_EXPORTER_OTLP_*`, and
  the sampler no longer overrides `OTEL_TRACES_SAMPLER`
- `MetricsAddr` may use port 0; `MetricsURL` reports the bound port
- Updated the root and example modules to Go 1.27.2, pinned the verify image
  to `golang-tooling:go1.27.2`, and moved the example Dockerfiles from
  `golang:1.24-alpine` to `golang:1.27.2-alpine`

### Changed

- `Config.TraceExporter` and `WithTraceExporter` now take an
  `sdktrace.SpanExporter` instead of `interface{}`, so a wrong type fails at
  compile time instead of in `Initialize`
- The resource now uses `deployment.environment.name` instead of the
  deprecated `deployment.environment`
- The resource now reports the OpenTelemetry Go SDK's real `telemetry.sdk.*`
  values instead of a hardcoded `otelx` / `1.0.0`
- `OTEL_RESOURCE_ATTRIBUTES` and `OTEL_SERVICE_NAME` are now read and override
  the values passed to `WithService`
- Upgraded `golang.org/x/net` to v0.60.0 for GO-2026-6603, GO-2026-6610,
  GO-2026-6611 and related HTTP/2 advisories
- Each `Telemetry` now serves metrics from its own Prometheus registry (with Go
  runtime and process collectors) instead of the global default registry

### Fixed

- The trace sampler now honors the parent span's sampling decision
  (`ParentBased(TraceIDRatioBased(rate))`), so a service no longer breaks up
  traces that an upstream service sampled
- `Initialize` now returns an error when the metrics port is already in use
  instead of printing it from a background goroutine
- Calling `Initialize` twice no longer breaks the `/metrics` endpoint with
  duplicate `target_info` errors
- `Initialize` no longer calls `zerolog.SetGlobalLevel`, which changed the level
  of every zerolog logger in the process
- `Shutdown` now joins errors with `errors.Join`, so `errors.Is` and `errors.As`
  work on the result

## [1.0.1] - 2026-08-21

### Added

- Docker-first verification build and GitHub Actions CI workflow
- Gin decorator-pattern example alongside the simple Gin example
- License and example documentation

### Changed

- Updated the root and example modules to Go 1.26.6
- Renamed the original Gin example to `gingonic-simple`
- Stopped tracking the generated `go.work.sum` file

### Fixed

- Addressed static-analysis findings in tests and test utilities

## [1.0.0] - 2025-12-16

### Added

- Core `otelx` library for unified OpenTelemetry initialization with
  `Initialize()`
- Configurable zerolog logging, Prometheus metrics, and OTLP gRPC tracing
- Service identity, signal, OTLP, and development or production configuration
  options
- OpenTelemetry resource setup, global provider registration, and W3C Trace
  Context and Baggage propagation
- `Telemetry.Shutdown()`, `Telemetry.Tracer()`, and `Telemetry.Meter()`
  helpers
- Gin middleware for trace-correlated request logging and context loggers
- Gin example with a Docker Compose observability stack
- Core library and middleware test coverage

## [0.0.1] - 2025-11-24

### Added

- Initial project setup

<!-- markdownlint-enable MD024 -->

[Unreleased]: https://github.com/twistingmercury/otelx/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/twistingmercury/otelx/releases/tag/v1.0.1
[1.0.0]: https://github.com/twistingmercury/otelx/releases/tag/v1.0.0
[0.0.1]: https://github.com/twistingmercury/otelx/releases/tag/v0.0.1
