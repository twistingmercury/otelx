package gin

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/twistingmercury/otelx"
)

// LoggingMiddleware creates middleware that provides trace-correlated logging
// for Gin HTTP handlers.
//
// The middleware performs the following:
//  1. Creates a trace-correlated logger using otelx.LoggerWithSpan
//  2. Stores the logger in gin.Context for retrieval via Logger or MustLogger
//  3. Logs request completion with status, duration, and trace IDs
//
// Configuration options can be provided to customize behavior:
//   - WithSkipPaths: Skip logging for specific paths (e.g., health checks)
//   - WithLogLevel: Set the log level for request completion logs
//   - WithRequestHeaders: Include specific request headers in logs
//   - WithCustomFields: Add custom fields to request completion logs
//
// Example:
//
//	r := gin.New()
//	r.Use(ginmw.LoggingMiddleware(tel,
//	    ginmw.WithSkipPaths("/health", "/ready"),
//	    ginmw.WithRequestHeaders("X-Request-ID"),
//	))
func LoggingMiddleware(tel *otelx.Telemetry, opts ...Option) gin.HandlerFunc {
	cfg := applyOptions(opts...)

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Create trace-correlated logger and store in context
		logger := otelx.LoggerWithSpan(tel.Logger, c.Request.Context())
		setLogger(c, logger)

		// Process request
		c.Next()

		// Skip logging for configured paths
		if cfg.skipPaths[path] {
			return
		}

		// Calculate request duration
		latency := time.Since(start)

		// Build log event
		var event *zerolog.Event
		status := c.Writer.Status()

		switch {
		case status >= 500:
			event = logger.WithLevel(zerolog.ErrorLevel)
		case status >= 400:
			event = logger.WithLevel(zerolog.WarnLevel)
		default:
			event = logger.WithLevel(cfg.logLevel)
		}

		// Add standard fields
		event.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Float64("latency_ms", float64(latency.Nanoseconds())/1e6).
			Str("client_ip", c.ClientIP())

		// Add query parameters if present
		if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
			event.Str("query", rawQuery)
		}

		// Add configured request headers
		for _, header := range cfg.requestHeaders {
			if value := c.GetHeader(header); value != "" {
				// Convert header name to snake_case for log field
				fieldName := headerToFieldName(header)
				event.Str(fieldName, value)
			}
		}

		// Add custom fields
		if cfg.customFields != nil {
			fields := cfg.customFields(c)
			for key, value := range fields {
				event.Interface(key, value)
			}
		}

		// Add error if present
		if len(c.Errors) > 0 {
			event.Str("error", c.Errors.String())
		}

		// Add response size
		event.Int("bytes", c.Writer.Size())

		// Log the request
		event.Msg("request completed")
	}
}

// CorrelationMiddleware creates middleware that stores a trace-correlated
// logger in gin.Context without logging request completion.
//
// Use this middleware when you need trace correlation but want to handle
// request logging separately (e.g., with another logging middleware or
// custom logging logic).
//
// The logger can be retrieved using Logger or MustLogger functions.
//
// Example:
//
//	r := gin.New()
//	r.Use(ginmw.CorrelationMiddleware(tel))
//
//	r.GET("/api/users", func(c *gin.Context) {
//	    logger := ginmw.Logger(c, tel)
//	    logger.Info().Msg("handling request")
//	})
func CorrelationMiddleware(tel *otelx.Telemetry) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create trace-correlated logger and store in context
		logger := otelx.LoggerWithSpan(tel.Logger, c.Request.Context())
		setLogger(c, logger)

		// Process request
		c.Next()
	}
}

// headerToFieldName converts an HTTP header name to a log field name.
// It converts to lowercase and replaces hyphens with underscores.
// Example: "X-Request-ID" becomes "x_request_id"
func headerToFieldName(header string) string {
	return strings.ToLower(strings.ReplaceAll(header, "-", "_"))
}
