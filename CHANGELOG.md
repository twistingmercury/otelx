# Changelog

<!-- markdownlint-disable MD024 -->

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Breaking

The library code moved under `src/`, so the module paths now include it:

| Before | After |
| --- | --- |
| `github.com/twistingmercury/otelx/v2` | `github.com/twistingmercury/otelx/src/v2` |
| `github.com/twistingmercury/otelx/middleware/gin/v2` | `github.com/twistingmercury/otelx/src/middleware/gin/v2` |

Update your imports and `go.mod` to the new paths. Releases are tagged
`src/v*` and `src/middleware/gin/v*`. The old `v2.0.0` tags still resolve under
the old paths but will not get further updates.

## [2.0.0] - 2026-10-14

### Breaking

The module path is now `github.com/twistingmercury/otelx/v2`, and
`Initialize` takes a `Config` struct instead of functional options. The Gin
middleware moved to its own module,
`github.com/twistingmercury/otelx/middleware/gin/v2`.

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

### Changed

- Exporter endpoint, headers, TLS and sampling come from the standard
  `OTEL_EXPORTER_OTLP_*` and `OTEL_TRACES_SAMPLER` variables instead of being
  overridden by otelx defaults
- `OTEL_SERVICE_NAME` and `OTEL_RESOURCE_ATTRIBUTES` are read and override the
  values in `Config`
- The resource uses `deployment.environment.name` instead of the deprecated
  `deployment.environment`, and reports the OpenTelemetry Go SDK's real
  `telemetry.sdk.*` values instead of a hardcoded `otelx` / `1.0.0`
- Each `Telemetry` serves metrics from its own Prometheus registry, with Go
  runtime and process collectors, instead of the global default registry
- `MetricsAddr` may use port 0, and `MetricsURL` reports the bound port
- Importing otelx no longer pulls in gin, sonic, quic-go or mongo-driver
- The Gin request log no longer includes the raw query string, which could
  carry tokens or personal data
- The README documents the actual API; the v1 README described options that
  never existed (`WithOTLPHeaders`, `WithoutMetrics`, `WithoutTracing`,
  `WithTraceSampler`)
- Updated the root and example modules to Go 1.27.2, pinned the verify image
  to `golang-tooling:go1.27.2`, and moved the example Dockerfiles from
  `golang:1.24-alpine` to `golang:1.27.2-alpine`
- The example docker-compose files set `OTEL_EXPORTER_OTLP_ENDPOINT` with an
  `http://` scheme, and `make run` defaults it to `http://localhost:4317`

### Fixed

- The trace sampler honors the parent span's sampling decision, so a service
  no longer breaks up traces that an upstream service sampled
- `Initialize` returns an error when the metrics port is already in use
  instead of printing it from a background goroutine
- Calling `Initialize` twice no longer breaks the `/metrics` endpoint with
  duplicate `target_info` errors
- `Initialize` no longer calls `zerolog.SetGlobalLevel`, which changed the
  level of every zerolog logger in the process
- `Shutdown` joins errors with `errors.Join`, so `errors.Is` and `errors.As`
  work on the result
- A custom trace exporter is typed as `sdktrace.SpanExporter`, so a wrong type
  fails at compile time instead of in `Initialize`

### Security

- Upgraded `golang.org/x/net` to v0.60.0 for GO-2026-6603, GO-2026-6610,
  GO-2026-6611 and related HTTP/2 advisories

## [1.0.2] - 2026-09-09

### Changed

- Updated the root and example modules, Go workspace, and verification build
  image to Go 1.27.1
- Refreshed dependencies in the root and example modules

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

[Unreleased]: https://github.com/twistingmercury/otelx/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/twistingmercury/otelx/compare/v1.0.2...v2.0.0
[1.0.2]: https://github.com/twistingmercury/otelx/releases/tag/v1.0.2
[1.0.1]: https://github.com/twistingmercury/otelx/releases/tag/v1.0.1
[1.0.0]: https://github.com/twistingmercury/otelx/releases/tag/v1.0.0
[0.0.1]: https://github.com/twistingmercury/otelx/releases/tag/v0.0.1
