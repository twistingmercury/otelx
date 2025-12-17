# Gin Simple Example

Demonstrates otelx integration with Gin using embedded telemetry in handlers.
Telemetry code (spans, metrics, logging) lives directly in handler methods
alongside business logic.

For running instructions, testing endpoints, and Docker setup, see the
[parent examples README](../README.md).

## Pattern Overview

The handler struct holds telemetry components:

```go
type Handler struct {
    tel          *otelx.Telemetry
    tracer       trace.Tracer
    uuidCounter  metric.Int64Counter
    durationHist metric.Float64Histogram
}
```

Each method creates spans, records metrics, and logs with trace correlation:

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

## Architecture

```text
main.go
   |
   v
otelx.Initialize() --> Telemetry
   |                      |
   v                      v
server.Initialize()  handlers.NewHandler(tel)
   |                      |
   v                      v
gin.Engine           Handler struct
   |                 (tracer, metrics)
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

## When to Use This Pattern

- Quick prototyping or small projects
- Team prefers seeing all code in one place
- Telemetry requirements are stable across endpoints

For the decorator alternative, see [gingonic-decorator](../gingonic-decorator/).
