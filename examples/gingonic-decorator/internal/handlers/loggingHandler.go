package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/twistingmercury/otelx"
	otelxgin "github.com/twistingmercury/otelx/middleware/gin"
)

// loggingHandler is a decorator that adds trace-correlated logging to a Handler.
type loggingHandler struct {
	tel          *otelx.Telemetry
	innerHandler Handler
}

// NewLoggingHandler creates a new logging decorator that wraps an inner Handler.
func NewLoggingHandler(tel *otelx.Telemetry, inner Handler) Handler {
	return &loggingHandler{
		tel:          tel,
		innerHandler: inner,
	}
}

// GetUUID delegates to the inner handler and logs with trace correlation.
func (lh *loggingHandler) GetUUID(c *gin.Context) {
	// Delegate to inner handler first
	lh.innerHandler.GetUUID(c)

	// Log with trace correlation after UUID is generated
	newUUID := c.GetString(uuidKey)
	logger := otelxgin.Logger(c, lh.tel)
	logger.Debug().Str("uuid", newUUID).Msg("generated UUID")
}
