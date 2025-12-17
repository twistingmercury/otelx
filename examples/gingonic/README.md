# Gin Web Application Example

Want to add observability to your Gin app without cluttering your business logic?
This example shows you how. We'll use the decorator pattern to keep your handlers
clean and testable while still getting all the telemetry goodness.

## What You Get

- **Structured Logging**: Zerolog with automatic trace correlation (so you can
  actually find your logs when debugging)
- **Prometheus Metrics**: A counter for UUID generation and a histogram for
  latency tracking
- **Distributed Tracing**: OTLP export with span attributes that tell you what's
  happening
- **Graceful Shutdown**: Everything cleans up properly when you stop the server

## The Decorator Pattern

Here's the key insight: we're separating "what the code does" from "how we
observe it." This might seem like extra work, but it pays off quickly when you
need to test things or add new telemetry without touching business logic.

Let's walk through the three pieces that make this work.

### Handler Interface

```go
type Handler interface {
GetUUID(c *gin.Context)
}
```

This interface is our contract - any handler that implements it can be wrapped
with telemetry. Simple, but powerful.

### ginHandler (Business Logic)

```go
type ginHandler struct{}

func (h *ginHandler) GetUUID(c *gin.Context) {
newUUID := uuid.New().String()
c.Set(uuidKey, newUUID)
c.JSON(http.StatusOK, gin.H{"uuid": newUUID})
}
```

Notice what's missing here? There's no logging, no metrics, no tracing - just
pure business logic. It generates a UUID and returns it as JSON. That's it.

Why does this matter? Because you can test this handler with zero mocking of
telemetry systems. Your unit tests stay simple and focused.

### observableHandler (Telemetry Decorator)

```go
type observableHandler struct {
tel          *otelx.Telemetry
tracer       trace.Tracer
uuidCounter  metric.Int64Counter
durationHist metric.Float64Histogram
innerHandler Handler
}
```

This is where the magic happens. The decorator wraps any `Handler` and adds all
the observability you need:

- Creates a child span for the operation
- Records metrics (counter increment, duration histogram)
- Adds span attributes with operation details
- Logs with trace correlation

The flow is straightforward: the decorator calls your inner handler for the
business logic, then enriches the request with telemetry data.

### Why Bother With This Pattern?

You might be wondering if this is overkill. Here's why we think it's worth it:

**Separation of Concerns**: Your business logic doesn't know telemetry exists.
Your telemetry doesn't care what the business logic does. They're completely
independent.

**Testability**: Want to test `ginHandler`? Just write simple unit tests. Want
to test `observableHandler`? Use mock handlers and verify the telemetry. Each
piece is easy to test in isolation.

**Single Responsibility**: Each struct does one thing well. When you need to add
a new feature, you'll know exactly where to put it - business logic goes in one
place, telemetry in another.

**Composability**: Here's the really cool part. You can wrap any handler with
observability. Need caching? Add a caching decorator. Rate limiting?
Authentication? Same pattern. You're building with Lego blocks, not writing
spaghetti.

## Architecture

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
server.NewServer()        |
   |                      |
   v                      v
Middleware Stack     handlers.NewObservableHandler()
   |                      |
   |   +------------------+
   |   |
   v   v
gin.Engine with routes
   |
   +---> GET /health -----> heartbeat.Handler
   |
   +---> GET /api/uuid ---> observableHandler
                                 |
                                 v
                           ginHandler (business logic)
```

**So what happens when a request hits /api/uuid?**

1. Request enters `gin.Engine`
2. `gin.Recovery()` middleware catches any panics (because things happen)
3. `otelgin.Middleware` creates the parent span
4. `loggingMiddleware` logs request completion with trace IDs
5. `observableHandler.GetUUID` creates a child span and records metrics
6. `ginHandler.GetUUID` does the actual work - generating the UUID
7. Response flows back through the middleware stack

## Running the Example

Ready to try it out? Start the server with:

```bash
make run
```

Or if you prefer to build the binary first:

```bash
make local
.out/uuid_api
```

You'll get two ports: 8080 for your API requests and 9090 for Prometheus metrics.

## Testing the Endpoints

### Health Check

```bash
curl http://localhost:8080/health
```

This gives you service health status along with dependency checks. Useful for
load balancers and monitoring systems.

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

| Endpoint  | Port | Description                         |
|-----------|------|-------------------------------------|
| /health   | 8080 | Health check with dependency status |
| /api/uuid | 8080 | Generate and return a new UUID      |
| /metrics  | 9090 | Prometheus metrics endpoint         |
