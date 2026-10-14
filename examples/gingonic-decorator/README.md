# Gin Decorator Pattern Example

Demonstrates otelx integration with Gin using the decorator pattern. Telemetry
concerns (tracing, metrics, logging) are separated into individual decorators
that wrap the core handler.

For running instructions, testing endpoints, and Docker setup, see the
[parent examples README](../README.md).

## Pattern Overview

Define a handler interface:

```go
type Handler interface {
    GetUUID(c *gin.Context)
}
```

The core handler contains only business logic:

```go
type ginHandler struct{}

func (h *ginHandler) GetUUID(c *gin.Context) {
    newUUID := uuid.New().String()
    c.Set(uuidKey, newUUID) // read by the logging decorator
    c.JSON(http.StatusOK, gin.H{"uuid": newUUID})
}
```

Each decorator wraps another handler and adds one concern:

```go
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

Compose decorators by wrapping each other. Tracing is outermost, so the
metrics and log line are recorded inside its span:

```go
ginHandler, err := handlers.NewGinHandler()
if err != nil {
    return err
}
loggingHandler := handlers.NewLoggingHandler(tel, ginHandler)
metricsHandler, err := handlers.NewMetricsHandler(tel, loggingHandler)
if err != nil {
    return err
}
tracingHandler := handlers.NewTracingHandler(tel, metricsHandler)

engine.GET("/api/uuid", tracingHandler.GetUUID)
```

## Request Flow

```text
Request
   |
   v
tracingHandler.GetUUID()    --> Starts span, puts it in the request context
   |
   v
metricsHandler.GetUUID()    --> Starts timing
   |
   v
loggingHandler.GetUUID()    --> Calls the core handler first
   |
   v
ginHandler.GetUUID()        --> Generates UUID, writes response
   |
   v
loggingHandler              --> Logs the UUID with trace and span IDs
   |
   v
metricsHandler              --> Records count and duration
   |
   v
tracingHandler              --> Ends span
   |
   v
Response
```

Separately, `LoggingMiddleware` from the otelx Gin module logs one line per
request, and `otelgin` creates the server span that `generate-uuid` nests
under.

## When to Use This Pattern

- You need to enable/disable telemetry features per endpoint
- Testing each concern in isolation is important
- Building reusable components across services
- Team values separation of concerns

For the simpler embedded approach, see [gingonic-simple](../gingonic-simple/).
