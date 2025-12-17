package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/twistingmercury/otelx"
	otelxgin "github.com/twistingmercury/otelx/middleware/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Handler handles HTTP requests with telemetry.
type Handler struct {
	tel          *otelx.Telemetry
	tracer       trace.Tracer
	uuidCounter  metric.Int64Counter
	durationHist metric.Float64Histogram
}

// NewHandler creates a new Handler.
func NewHandler(tel *otelx.Telemetry) (*Handler, error) {
	meter := tel.Meter("uuid-api")

	uuidCounter, err := meter.Int64Counter("uuid_generated_total",
		metric.WithDescription("Total number of UUIDs generated"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create counter: %w", err)
	}

	durationHist, err := meter.Float64Histogram("uuid_generation_duration_seconds",
		metric.WithDescription("Duration of UUID generation"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create histogram: %w", err)
	}

	return &Handler{
		tel:          tel,
		tracer:       tel.Tracer("uuid-api"),
		uuidCounter:  uuidCounter,
		durationHist: durationHist,
	}, nil
}

// GetUUID generates and returns a UUID with full telemetry.
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
