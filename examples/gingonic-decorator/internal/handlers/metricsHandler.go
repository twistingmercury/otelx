package handlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twistingmercury/otelx"
	"go.opentelemetry.io/otel/metric"
)

// metricsHandler is a decorator that adds metrics to a Handler.
type metricsHandler struct {
	uuidCounter  metric.Int64Counter
	durationHist metric.Float64Histogram
	innerHandler Handler
}

// NewMetricsHandler creates a new metrics decorator that wraps an inner Handler.
func NewMetricsHandler(tel *otelx.Telemetry, inner Handler) (Handler, error) {
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

	return &metricsHandler{
		uuidCounter:  uuidCounter,
		durationHist: durationHist,
		innerHandler: inner,
	}, nil
}

// GetUUID records metrics around the inner handler call.
func (mh *metricsHandler) GetUUID(c *gin.Context) {
	ctx := c.Request.Context()

	// Record timing
	start := time.Now()
	mh.innerHandler.GetUUID(c)
	duration := time.Since(start).Seconds()

	// Record metrics
	mh.uuidCounter.Add(ctx, 1)
	mh.durationHist.Record(ctx, duration)
}
