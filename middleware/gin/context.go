package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/twistingmercury/otelx"
)

// loggerKey is the key used to store the logger in gin.Context.
const loggerKey = "otelx.logger"

// Logger retrieves the trace-correlated logger from gin.Context.
// If the middleware has not been applied or the logger is not found,
// it falls back to the logger from the provided Telemetry instance.
//
// This function is safe to call even when the middleware is not applied,
// making it useful for shared code that may or may not be running in a
// middleware context.
//
// Example:
//
//	func handler(c *gin.Context) {
//	    logger := ginmw.Logger(c, tel)
//	    logger.Info().Msg("processing request")
//	}
func Logger(c *gin.Context, tel *otelx.Telemetry) zerolog.Logger {
	if logger, exists := c.Get(loggerKey); exists {
		if l, ok := logger.(zerolog.Logger); ok {
			return l
		}
	}

	// Fallback to telemetry logger with span context if available
	if tel != nil {
		return otelx.LoggerWithSpan(tel.Logger, c.Request.Context())
	}

	return zerolog.Nop()
}

// MustLogger retrieves the trace-correlated logger from gin.Context.
// It panics if the logger is not found in the context.
//
// Use this function only when you are certain the middleware has been applied.
// For safer usage that provides a fallback, use Logger instead.
//
// Example:
//
//	func handler(c *gin.Context) {
//	    logger := ginmw.MustLogger(c)
//	    logger.Info().Msg("processing request")
//	}
func MustLogger(c *gin.Context) zerolog.Logger {
	logger, exists := c.Get(loggerKey)
	if !exists {
		panic("otelx gin middleware: logger not found in context - ensure middleware is applied")
	}

	l, ok := logger.(zerolog.Logger)
	if !ok {
		panic("otelx gin middleware: invalid logger type in context")
	}

	return l
}

// setLogger stores the logger in gin.Context.
func setLogger(c *gin.Context, logger zerolog.Logger) {
	c.Set(loggerKey, logger)
}
