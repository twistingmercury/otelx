// Package gin provides OpenTelemetry-integrated middleware for the Gin web framework.
//
// This package provides middleware that automatically correlates logs with traces
// by injecting trace and span IDs into the zerolog logger context. It integrates
// with the otelx package to provide a unified telemetry experience.
//
// # Basic Usage
//
// The simplest way to use this package is with LoggingMiddleware, which creates
// a trace-correlated logger, stores it in the gin.Context, and logs request
// completion with status, duration, and trace IDs:
//
//	import (
//	    "github.com/gin-gonic/gin"
//	    otelx "github.com/twistingmercury/otelx/src/v2"
//	    ginmw "github.com/twistingmercury/otelx/src/middleware/gin/v2"
//	)
//
//	func main() {
//	    tel, err := otelx.Initialize(ctx, otelx.Config{
//	        ServiceName: "my-service",
//	        Tracing:     true,
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer tel.Shutdown(ctx)
//
//	    r := gin.New()
//	    r.Use(ginmw.LoggingMiddleware(tel))
//
//	    r.GET("/api/users", func(c *gin.Context) {
//	        logger := ginmw.Logger(c, tel)
//	        logger.Info().Msg("handling users request")
//	        // ... handler logic
//	    })
//	}
//
// # Retrieving the Logger
//
// Within your handlers, retrieve the trace-correlated logger using the Logger
// or MustLogger functions:
//
//	// Logger returns the trace-correlated logger, falling back to tel.Logger
//	logger := ginmw.Logger(c, tel)
//
//	// MustLogger panics if the logger is not found (use when middleware is guaranteed)
//	logger := ginmw.MustLogger(c)
//
// # Configuration Options
//
// The middleware can be configured using functional options:
//
//	r.Use(ginmw.LoggingMiddleware(tel,
//	    ginmw.WithSkipPaths("/health", "/ready", "/metrics"),
//	    ginmw.WithLogLevel(zerolog.DebugLevel),
//	    ginmw.WithRequestHeaders("X-Request-ID", "X-Correlation-ID"),
//	    ginmw.WithCustomFields(func(c *gin.Context) map[string]interface{} {
//	        return map[string]interface{}{
//	            "user_id": c.GetString("user_id"),
//	        }
//	    }),
//	))
//
// # Correlation-Only Mode
//
// If you only need trace correlation without request logging (e.g., when using
// another logging middleware), use CorrelationMiddleware:
//
//	r.Use(ginmw.CorrelationMiddleware(tel))
//
// This creates and stores the trace-correlated logger without logging request
// completion details.
//
// # Log Fields
//
// LoggingMiddleware automatically logs the following fields on request completion:
//   - method: HTTP method (GET, POST, etc.)
//   - path: Request path
//   - status: HTTP status code
//   - latency_ms: Request duration in milliseconds
//   - client_ip: Client IP address
//   - bytes: Response body size in bytes
//   - error: Errors attached to the gin.Context, when there are any
//   - trace_id: OpenTelemetry trace ID (if tracing is active)
//   - span_id: OpenTelemetry span ID (if tracing is active)
//   - Any configured request headers
//   - Any custom fields from WithCustomFields
//
// The raw query string is never logged, because it can carry tokens or
// personal data.
package gin
