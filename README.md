# otelx

> **Maturity Level**: Emerging - Initial development, API subject to change

A Go library for unified OpenTelemetry initialization with a single function call.

## Usage

Initialize all telemetry signals with one call:

```go
package main

import (
    "context"
    "log"

    "github.com/twistingmercury/otelx"
)

func main() {
    ctx := context.Background()

    tel, err := otelx.Initialize(ctx,
        otelx.WithService("my-service", "1.0.0", "production"),
        otelx.WithMetrics(9090),
        otelx.WithTracing(),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer tel.Shutdown(ctx)

    // Use the initialized components
    tel.Logger.Info().Msg("service started")

    // Create metrics
    meter := tel.MeterProvider.Meter("my-service")
    counter, _ := meter.Int64Counter("requests_total")
    counter.Add(ctx, 1)

    // Create traces
    tracer := tel.TracerProvider.Tracer("my-service")
    ctx, span := tracer.Start(ctx, "operation")
    defer span.End()
}
```

### Configuration Options

#### Service Identity (Required)

```go
otelx.WithService("service-name", "1.0.0", "production")
```

#### Logging

```go
otelx.WithLogLevel(zerolog.DebugLevel)  // Default: Info
otelx.WithLogWriter(os.Stderr)          // Default: os.Stdout
otelx.WithoutLogging()                  // Disable logging
```

#### Metrics

```go
otelx.WithMetrics(9090)                 // Enable Prometheus on port
otelx.WithMetricsPath("/custom/metrics") // Default: /metrics
otelx.WithoutMetrics()                  // Disable metrics
```

#### Tracing

```go
otelx.WithTracing()                     // Enable OTLP tracing
otelx.WithTraceSampler(sdktrace.AlwaysSample())
otelx.WithTraceSampleRate(0.1)          // 10% sampling
otelx.WithoutTracing()                  // Disable tracing
```

#### OTLP Configuration

```go
otelx.WithOTLPEndpoint("collector.example.com:4317") // Default: localhost:4317
otelx.WithOTLPInsecure()                             // Disable TLS
otelx.WithOTLPHeaders(map[string]string{"Authorization": "Bearer token"})
```

#### Presets

```go
otelx.WithAllSignals()          // Metrics (9090) + tracing
otelx.WithDevelopmentDefaults() // Debug, pretty logging, insecure OTLP
otelx.WithProductionDefaults()  // Info, JSON logging, TLS, 10% sampling
```

### The Telemetry Struct

`Initialize()` returns a `*Telemetry` struct:

```go
type Telemetry struct {
    Logger         zerolog.Logger
    MeterProvider  *sdkmetric.MeterProvider
    TracerProvider *sdktrace.TracerProvider
}
```

Call `Shutdown()` when your application exits to flush pending telemetry data.

## How it works

otelx wraps the official OpenTelemetry Go SDK to provide a unified
initialization experience. Instead of configuring logging, metrics, and tracing
separately, you call `Initialize()` once with functional options.

**Logging**: Uses [zerolog](https://github.com/rs/zerolog) with automatic trace
correlation via
[otelzerolog](https://github.com/agoda-com/opentelemetry-go/tree/main/otelzerolog).
Log entries automatically include trace and span IDs when a trace context is
active.

**Metrics**: Uses the official OpenTelemetry Prometheus exporter. When enabled,
otelx starts an HTTP server on the specified port exposing metrics at the
configured path.

**Tracing**: Uses the OTLP gRPC exporter to send traces to an OpenTelemetry
Collector or compatible backend.

The functional options pattern lets you enable only the signals you need and
customize each component independently.

## Key Considerations

- **WithService is required**: You must provide service name, version, and
  environment for proper telemetry identification
- **Shutdown is essential**: Always call `tel.Shutdown(ctx)` to flush pending
  telemetry before your application exits
- **Metrics port binding**: When metrics are enabled, otelx binds to the
  specified port; ensure it is available
- **OTLP endpoint**: Tracing requires an OpenTelemetry Collector or compatible
  endpoint; configure `WithOTLPEndpoint()` for non-local deployments
- **TLS by default**: Production tracing uses TLS; use `WithOTLPInsecure()` only
  for local development

## Development Considerations

### Quick Start

```bash
go mod tidy
go build ./...
```

### Building and running

Build the library:

```bash
go build ./...
```

Run examples from the `_examples/` directory:

```bash
go run ./_examples/basic
```

### Testing

Run all tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

### Versioning

This project uses git tag-based semantic versioning. Version tags follow the
format `vX.Y.Z`.

Import the module:

```go
import "github.com/twistingmercury/otelx"
```

## Known Issues

### Go Standard Library Vulnerabilities

The following vulnerabilities appear in govulncheck but are **not in otelx code**:

- **GO-2025-4175**: Improper DNS name constraint handling in wildcard
  verification (crypto/x509)
- **GO-2025-4155**: Excessive resource consumption in certificate validation
  error strings (crypto/x509)

These are in the Go standard library and are pulled in transitively via
prometheus/promhttp -> crypto/tls -> crypto/x509.

**Impact**: Minimal for typical otelx usage. Most deployments use plain HTTP for
Prometheus metrics endpoints and in-cluster OTLP without TLS certificate
validation concerns.

**Resolution**: Upgrade to Go 1.25.5 or later. Users who enable TLS for metrics
endpoints or OTLP connections should prioritize this upgrade.

## License

MIT License. See [LICENSE](LICENSE) for details.
