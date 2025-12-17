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

Compose decorators by wrapping each other:

```go
ginHandler, _ := handlers.NewGinHandler()
tracingHandler, _ := handlers.NewTracingHandler(tel, ginHandler)
metricsHandler, _ := handlers.NewMetricsHandler(tel, tracingHandler)
loggingHandler, _ := handlers.NewLoggingHandler(tel, metricsHandler)

engine.GET("/api/uuid", loggingHandler.GetUUID)
```

## Request Flow

```text
Request
   |
   v
loggingHandler.GetUUID()    --> Logs request start
   |
   v
metricsHandler.GetUUID()    --> Starts timing
   |
   v
tracingHandler.GetUUID()    --> Creates span
   |
   v
ginHandler.GetUUID()        --> Generates UUID, writes response
   |
   v
tracingHandler              --> Ends span
   |
   v
metricsHandler              --> Records duration
   |
   v
loggingHandler              --> Logs completion with trace ID
   |
   v
Response
```

## Tradeoffs: Combined vs Separate Decorators

This example shows both approaches for educational purposes.

| Aspect | Single Combined | Three Separate |
|--------|-----------------|----------------|
| Single Responsibility | Violates (3 concerns) | Each does one thing |
| Composability | All-or-nothing | Mix and match |
| Simplicity | Easier to use | More setup code |
| Testability | Test all together | Test each independently |

## When to Use This Pattern

- You need to enable/disable telemetry features per endpoint
- Testing each concern in isolation is important
- Building reusable components across services
- Team values separation of concerns

For the simpler embedded approach, see [gingonic-simple](../gingonic-simple/).
