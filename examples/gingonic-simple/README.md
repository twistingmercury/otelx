# Gin Web Application Example

Want to add observability to your Gin app? This example shows you how to use otelx
for structured logging, metrics, and distributed tracing with minimal code.

## What You Get

- **Structured Logging**: Zerolog with automatic trace correlation (so you can
  actually find your logs when debugging)
- **Prometheus Metrics**: A counter for UUID generation and a histogram for
  latency tracking
- **Distributed Tracing**: OTLP export with span attributes that tell you what's
  happening
- **Graceful Shutdown**: Everything cleans up properly when you stop the server

## Architecture

The example has three files with clear responsibilities:

```text
main.go
   |
   v
otelx.Initialize() --> Telemetry
   |                      |
   |          +-----------+------------+
   |          |           |            |
   |       Logger    MeterProvider  TracerProvider
   |          |           |            |
   v          +-----------+------------+
server.Initialize()       |
   |                      v
   |              handlers.NewHandler(tel)
   |                      |
   v                      v
gin.Engine         Handler struct
   |                (tracer, metrics)
   |
   +---> Middleware: Recovery, otelgin, LoggingMiddleware
   |
   +---> GET /health -----> inline health check
   |
   +---> GET /api/uuid ---> Handler.GetUUID
                                 |
                                 +-> Creates span
                                 +-> Generates UUID
                                 +-> Records metrics
                                 +-> Logs with trace ID
                                 +-> Returns JSON
```

## How It Works

### 1. Initialize Telemetry (main.go)

```go
tel, err := otelx.Initialize(ctx,
    otelx.WithService("uuid-api", "1.0.0", "development"),
    otelx.WithDevelopmentDefaults(),
    otelx.WithMetrics(9090),
    otelx.WithTracing(),
    otelx.WithOTLPEndpoint(endpoint),
)
```

This gives you a `Telemetry` struct containing a logger, tracer provider, and
meter provider. Pass it to your handlers and they have everything they need.

### 2. Set Up Middleware (server.go)

```go
engine.Use(gin.Recovery())
engine.Use(otelgin.Middleware("uuid-api"))
engine.Use(otelxgin.LoggingMiddleware(tel,
    otelxgin.WithSkipPaths("/health"),
))
```

The middleware stack handles panic recovery, creates parent spans for each
request, and logs request completion with trace IDs.

### 3. Use Telemetry in Handlers (handlers.go)

```go
type Handler struct {
    tel          *otelx.Telemetry
    tracer       trace.Tracer
    uuidCounter  metric.Int64Counter
    durationHist metric.Float64Histogram
}

func (h *Handler) GetUUID(c *gin.Context) {
    ctx := c.Request.Context()

    // Create span
    ctx, span := h.tracer.Start(ctx, "generate-uuid",
        trace.WithSpanKind(trace.SpanKindInternal),
    )
    defer span.End()

    // Generate UUID with timing
    start := time.Now()
    newUUID := uuid.New().String()
    duration := time.Since(start).Seconds()

    // Record metrics
    h.uuidCounter.Add(ctx, 1)
    h.durationHist.Record(ctx, duration)

    // Add span attributes
    span.SetAttributes(attribute.String("uuid.value", newUUID))

    // Log with trace correlation
    logger := otelxgin.Logger(c, h.tel)
    logger.Debug().Str("uuid", newUUID).Msg("generated UUID")

    c.JSON(http.StatusOK, gin.H{"uuid": newUUID})
}
```

All the telemetry lives right alongside your business logic. Create spans where
you need visibility, record metrics where they matter, and log with automatic
trace correlation.

## Running the Example

Start the server with:

```bash
make run
```

Or if you prefer to build the binary first:

```bash
make local
.out/uuid_api
```

You'll get two ports: 8080 for your API requests and 9090 for Prometheus metrics.

## Running with Docker (Full Observability Stack)

Want to see the complete picture with trace visualization and metrics querying? The
Docker Compose setup gives you the full observability stack.

```bash
docker compose up --build
```

This spins up four services:

| Service       | URL                             | Purpose                    |
|---------------|---------------------------------|----------------------------|
| API           | <http://localhost:8080>         | Gin UUID service           |
| Metrics       | <http://localhost:9090/metrics> | Prometheus metrics         |
| Prometheus UI | <http://localhost:9091>         | Query and graph metrics    |
| Jaeger UI     | <http://localhost:16686>        | Visualize traces           |

To see everything working together:

1. Generate some UUIDs: `curl http://localhost:8080/api/uuid`
2. Open Jaeger at <http://localhost:16686>, select "uuid-api" service, and click
   "Find Traces" - you'll see the full request flow with spans and timing
3. Open Prometheus at <http://localhost:9091> and query `uuid_generated_total`
   to see your counter incrementing

The OpenTelemetry Collector sits between your app and Jaeger, receiving traces via
OTLP and forwarding them. This is the production pattern - your app talks to the
collector, not directly to backends.

## Testing the Endpoints

### Health Check

```bash
curl http://localhost:8080/health
```

Returns `{"status": "ok"}` when the service is running.

### UUID Generation

```bash
curl http://localhost:8080/api/uuid
```

You'll get back something like:

```json
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Prometheus Metrics

```bash
curl http://localhost:9090/metrics
```

Here you can see all your Prometheus metrics, including:

- `uuid_generated_total`: How many UUIDs have been generated
- `uuid_generation_duration_seconds`: How long each generation took (histogram)

## Endpoints

| Endpoint  | Port | Description                    |
|-----------|------|--------------------------------|
| /health   | 8080 | Health check                   |
| /api/uuid | 8080 | Generate and return a new UUID |
| /metrics  | 9090 | Prometheus metrics endpoint    |
