package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/twistingmercury/otelx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// tracingHandler is a decorator that adds tracing to a Handler.
type tracingHandler struct {
	tracer       trace.Tracer
	innerHandler Handler
}

// NewTracingHandler creates a new tracing decorator that wraps an inner Handler.
func NewTracingHandler(tel *otelx.Telemetry, inner Handler) Handler {
	return &tracingHandler{
		tracer:       tel.Tracer("uuid-api"),
		innerHandler: inner,
	}
}

// GetUUID creates a span, delegates to the inner handler, and sets span attributes.
func (th *tracingHandler) GetUUID(c *gin.Context) {
	ctx := c.Request.Context()

	// Create span
	ctx, span := th.tracer.Start(ctx, "generate-uuid",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	// Update request context so inner handlers can access the span
	c.Request = c.Request.WithContext(ctx)

	// Delegate to inner handler
	th.innerHandler.GetUUID(c)

	// Add span attributes after inner handler has set the UUID
	newUUID := c.GetString(uuidKey)
	span.SetAttributes(attribute.String("uuid.value", newUUID))
}
