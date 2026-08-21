# otelx Gin Examples

Two examples demonstrating otelx integration with the Gin web framework. Both
implement the same UUID generation API with full observability, but use
different architectural patterns for organizing telemetry code.

| Example | Pattern | Best For |
|---------|---------|----------|
| [gingonic-simple](gingonic-simple/) | Embedded telemetry | Quick setup, smaller projects |
| [gingonic-decorator](gingonic-decorator/) | Decorator chain | Separation of concerns, testability |

## What You Get

Both examples provide:

- **Structured Logging**: Zerolog with automatic trace correlation
- **Prometheus Metrics**: Counter for UUID generation, histogram for latency
- **Distributed Tracing**: OTLP export with span attributes
- **Graceful Shutdown**: Clean telemetry flush on server stop

## Prerequisites

- Go 1.26.6+
- Docker and Docker Compose (for full observability stack)

## Running an Example

Each example directory contains a Makefile for local development.

### Local Development

```bash
cd gingonic-simple  # or gingonic-decorator
make run
```

Or build first:

```bash
make local
.out/uuid_api
```

Ports: 8080 (API), 9090 (Prometheus metrics).

### Full Observability Stack

Run with Docker Compose for the complete observability experience:

```bash
cd gingonic-simple  # or gingonic-decorator
docker compose up --build
```

Services started:

| Service       | URL                             | Purpose                 |
|---------------|---------------------------------|-------------------------|
| API           | <http://localhost:8080>         | Gin UUID service        |
| Metrics       | <http://localhost:9090/metrics> | Prometheus metrics      |
| Prometheus UI | <http://localhost:9091>         | Query and graph metrics |
| Jaeger UI     | <http://localhost:16686>        | Visualize traces        |

The OpenTelemetry Collector sits between your app and Jaeger, receiving traces
via OTLP and forwarding them. This is the production pattern.

## Testing the API

### Health Check

```bash
curl http://localhost:8080/health
# {"status": "ok"}
```

### UUID Generation

```bash
curl http://localhost:8080/api/uuid
# {"uuid": "550e8400-e29b-41d4-a716-446655440000"}
```

### View Telemetry

1. Generate some UUIDs: `curl http://localhost:8080/api/uuid`
2. Open Jaeger at <http://localhost:16686>, select "uuid-api" service, click
   "Find Traces"
3. Open Prometheus at <http://localhost:9091> and query `uuid_generated_total`

### Prometheus Metrics

```bash
curl http://localhost:9090/metrics
```

Available metrics:

- `uuid_generated_total`: Count of generated UUIDs
- `uuid_generation_duration_seconds`: Generation latency histogram

## Endpoints

| Endpoint  | Port | Description                    |
|-----------|------|--------------------------------|
| /health   | 8080 | Health check                   |
| /api/uuid | 8080 | Generate and return a new UUID |
| /metrics  | 9090 | Prometheus metrics endpoint    |

## Simple Pattern

The [simple example](gingonic-simple/) embeds telemetry directly in handler
methods. The handler struct holds the tracer and metrics, and each method
creates spans, records metrics, and logs.

```go
func (h *Handler) GetUUID(c *gin.Context) {
    ctx, span := h.tracer.Start(c.Request.Context(), "generate-uuid")
    defer span.End()

    start := time.Now()
    newUUID := uuid.New().String()
    duration := time.Since(start).Seconds()

    h.uuidCounter.Add(ctx, 1)
    h.durationHist.Record(ctx, duration)
    span.SetAttributes(attribute.String("uuid.value", newUUID))

    logger := otelxgin.Logger(c, h.tel)
    logger.Debug().Str("uuid", newUUID).Msg("generated UUID")

    c.JSON(http.StatusOK, gin.H{"uuid": newUUID})
}
```

Telemetry and business logic are interleaved in the same function.

## Decorator Pattern

The [decorator example](gingonic-decorator/) separates concerns using the
decorator pattern. Each decorator wraps a handler and adds one type of
telemetry behavior.

```go
type Handler interface {
    GetUUID(c *gin.Context)
}

// Core handler - just business logic
type ginHandler struct{}

func (h *ginHandler) GetUUID(c *gin.Context) {
    newUUID := uuid.New().String()
    c.JSON(http.StatusOK, gin.H{"uuid": newUUID})
}

// Tracing decorator
type tracingHandler struct {
    tracer       trace.Tracer
    innerHandler Handler
}

func (th *tracingHandler) GetUUID(c *gin.Context) {
    ctx, span := th.tracer.Start(c.Request.Context(), "generate-uuid")
    defer span.End()
    c.Request = c.Request.WithContext(ctx)
    th.innerHandler.GetUUID(c)
}
```

Decorators compose by wrapping each other:

```go
ginHandler, _ := handlers.NewGinHandler()
tracingHandler, _ := handlers.NewTracingHandler(tel, ginHandler)
metricsHandler, _ := handlers.NewMetricsHandler(tel, tracingHandler)
loggingHandler, _ := handlers.NewLoggingHandler(tel, metricsHandler)

engine.GET("/api/uuid", loggingHandler.GetUUID)
```

Request flow:

```text
Request --> Logging --> Metrics --> Tracing --> Core Handler --> Response
```

## Pattern Comparison

| Aspect | Simple | Decorator |
|--------|--------|-----------|
| Code organization | Single file, interleaved | Separate concerns |
| Setup complexity | Low | Higher |
| Testing | Test all telemetry together | Test each concern independently |
| Flexibility | All-or-nothing | Mix and match decorators |
| Readability | Business logic harder to find | Core logic is isolated |
| Maintenance | Changes affect one file | Changes isolated to decorator |

### When to Use Each

**Simple pattern works well when:**

- Quick prototyping or small projects
- Team prefers seeing all code in one place
- Telemetry requirements are stable and uniform

**Decorator pattern works well when:**

- You need to enable/disable telemetry features per endpoint
- Testing each concern in isolation is important
- Building reusable components across services
- Team values separation of concerns
