# Changelog

<!-- markdownlint-disable MD024 -->

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
