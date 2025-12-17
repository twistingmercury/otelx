# Plan: otelx Standalone Project

## Overview

Create a new standalone Go project `otelx` at `/Users/doublej/dev/otelx` as the spiritual successor to the deprecated `github.com/twistingmercury/telemetry/v2`.

**Module Path:** `github.com/twistingmercury/otelx`
**Import Path:** `github.com/twistingmercury/otelx`

---

## Project Structure

```text
/Users/doublej/dev/otelx/
├── .github/
│   └── workflows/
│       ├── ci.yaml              # Runs build/build.sh
│       └── release.yaml         # Automated releases from tags
├── .golangci.yaml               # Linter config
├── LICENSE                      # MIT
├── README.md                    # Documentation
├── CHANGELOG.md                 # Version history
├── CONTRIBUTING.md              # Contribution guide
├── MIGRATION.md                 # Migration from telemetry/v2
├── Makefile                     # Build targets (build, test, e2e-*, help)
├── go.mod / go.sum
├── doc.go                       # Package docs
├── otelx.go                     # Initialize(), Telemetry struct
├── config.go                    # Config struct, defaults, validation
├── options.go                   # All With*() options
├── logging.go                   # otelzerolog bridge setup
├── metrics.go                   # Prometheus exporter + HTTP server
├── tracing.go                   # OTLP gRPC exporter
├── resource.go                  # OTel resource creation
├── *_test.go                    # Unit tests for each file
├── example_test.go              # Godoc examples
├── build/
│   └── build.sh                 # Main build script (unit test, build, e2e)
├── scripts/
│   └── print.sh                 # Shell helper functions
├── tests/                       # E2E tests (handled by go-e2e-test-agent)
│   └── ...
├── internal/
│   └── testutil/
│       └── testutil.go          # Test helpers
└── _examples/
    ├── basic/main.go
    ├── development/main.go
    ├── production/main.go
    └── custom/main.go
```

---

## Core API

```go
import "github.com/twistingmercury/otelx"

// Telemetry holds initialized components
type Telemetry struct {
    Logger         zerolog.Logger
    MeterProvider  *sdkmetric.MeterProvider
    TracerProvider *sdktrace.TracerProvider
}

func Initialize(ctx context.Context, opts ...Option) (*Telemetry, error)
func (t *Telemetry) Shutdown(ctx context.Context) error
```

### Options

| Option                            | Description                      |
| --------------------------------- | -------------------------------- |
| `WithService(name, version, env)` | **Required** - service identity  |
| `WithLogLevel(level)`             | Log level (default: Info)        |
| `WithLogWriter(w)`                | Log output (default: os.Stdout)  |
| `WithoutLogging()`                | Disable logging                  |
| `WithMetrics(port)`               | Enable Prometheus on port        |
| `WithMetricsPath(path)`           | Metrics path (default: /metrics) |
| `WithTracing()`                   | Enable OTLP tracing              |
| `WithTraceSampleRate(rate)`       | Sampling rate 0.0-1.0            |
| `WithTraceExporter(exp)`          | Custom exporter                  |
| `WithOTLPEndpoint(endpoint)`      | Collector endpoint               |
| `WithOTLPInsecure()`              | Disable TLS                      |
| `WithAllSignals()`                | Enable metrics(9090) + tracing   |
| `WithDevelopmentDefaults()`       | Debug, pretty, insecure          |
| `WithProductionDefaults()`        | Info, JSON, TLS, 10% sampling    |

---

## Dependencies

```go
require (
    github.com/rs/zerolog v1.33.0
    github.com/agoda-com/otelzerolog v0.1.0
    go.opentelemetry.io/otel v1.38.0
    go.opentelemetry.io/otel/sdk v1.38.0
    go.opentelemetry.io/otel/sdk/metric v1.38.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0
    go.opentelemetry.io/otel/exporters/prometheus v0.60.0
    github.com/prometheus/client_golang v1.23.0
    github.com/stretchr/testify v1.11.1
)
```

---

## Implementation Phases

### Phase 1: Foundation

1. Create directory: `/Users/doublej/dev/otelx/`
2. Initialize git repo
3. `go.mod`, `.gitignore`
4. `doc.go`, `config.go`, `resource.go`
5. `internal/testutil/testutil.go`

### Phase 2: Telemetry Signals

6. `logging.go` + `logging_test.go`
7. `metrics.go` + `metrics_test.go`
8. `tracing.go` + `tracing_test.go`

### Phase 3: Main API

9. `options.go` + `options_test.go`
10. `otelx.go` + `otelx_test.go`
11. `example_test.go`

### Phase 4: Build Infrastructure

12. `scripts/print.sh` - shell helper functions
13. `build/build.sh` - main build script
14. `Makefile` - build targets

### Phase 5: Documentation & CI

15. `README.md`, `CHANGELOG.md`, `MIGRATION.md`, `CONTRIBUTING.md`, `LICENSE`
16. `.golangci.yaml`, `.github/workflows/ci.yaml`, `.github/workflows/release.yaml`

### Phase 6: Examples

17. `_examples/basic/main.go`
18. `_examples/development/main.go`
19. `_examples/production/main.go`
20. `_examples/custom/main.go`

### Phase 7: E2E Tests (separate)

21. `tests/` - handled by go-e2e-test-agent

### Phase 8: Update Old Telemetry Project

22. Update `/Users/doublej/dev/telemetry/MIGRATION.md` to point to otelx
23. Ensure deprecation notice in telemetry README points to otelx

---

## Key Design Decisions

1. **Flat structure** - Code at root for simple `import "github.com/twistingmercury/otelx"`
2. **Return struct + set globals** - Best of both worlds
3. **Official otelzerolog bridge** - `github.com/agoda-com/otelzerolog`
4. **Options return error** - Early validation at Initialize() time
5. **No framework dependencies** - Works with any HTTP framework

---

## Delegation Plan

| Phase | Agent               | Task                                                       |
| ----- | ------------------- | ---------------------------------------------------------- |
| 1-3   | go-software-agent   | Create all Go source files + unit tests                    |
| 4     | shell-script-agent  | Create `scripts/print.sh`, `build/build.sh`                |
| 4     | shell-script-agent  | Create `Makefile`                                          |
| 5     | documentation-agent | Create README, CHANGELOG, MIGRATION, CONTRIBUTING, LICENSE |
| 5     | go-devops-agent     | Create `.golangci.yaml`, CI/CD workflows                   |
| 6     | go-software-agent   | Create example applications                                |
| 7     | go-e2e-test-agent   | Create `tests/` directory with e2e tests                   |
| 8     | documentation-agent | Update old telemetry MIGRATION.md                          |
