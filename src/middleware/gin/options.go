package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Option is a function that configures the middleware.
type Option func(*config)

// config holds the middleware configuration.
type config struct {
	// skipPaths contains paths that should skip logging.
	skipPaths map[string]bool

	// logLevel is the log level for request completion logs.
	logLevel zerolog.Level

	// requestHeaders contains header names to include in logs.
	requestHeaders []string

	// customFields is a function that returns additional fields to log.
	customFields func(*gin.Context) map[string]interface{}
}

// defaultConfig returns the default middleware configuration.
func defaultConfig() *config {
	return &config{
		skipPaths:      make(map[string]bool),
		logLevel:       zerolog.InfoLevel,
		requestHeaders: nil,
		customFields:   nil,
	}
}

// applyOptions applies the given options to the config.
func applyOptions(opts ...Option) *config {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// WithSkipPaths configures paths that should skip request logging.
// The logger will still be stored in context for these paths, but no
// request completion log will be emitted.
//
// Example:
//
//	ginmw.LoggingMiddleware(tel,
//	    ginmw.WithSkipPaths("/health", "/ready", "/metrics"),
//	)
func WithSkipPaths(paths ...string) Option {
	return func(c *config) {
		for _, path := range paths {
			c.skipPaths[path] = true
		}
	}
}

// WithLogLevel sets the log level for request completion logs.
// Default is zerolog.InfoLevel.
//
// Example:
//
//	ginmw.LoggingMiddleware(tel,
//	    ginmw.WithLogLevel(zerolog.DebugLevel),
//	)
func WithLogLevel(level zerolog.Level) Option {
	return func(c *config) {
		c.logLevel = level
	}
}

// WithRequestHeaders configures which request headers to include in logs.
// Header values will be logged with the header name converted to lowercase
// with underscores replacing hyphens (e.g., "X-Request-ID" becomes "x_request_id").
//
// Example:
//
//	ginmw.LoggingMiddleware(tel,
//	    ginmw.WithRequestHeaders("X-Request-ID", "X-Correlation-ID"),
//	)
func WithRequestHeaders(headers ...string) Option {
	return func(c *config) {
		c.requestHeaders = append(c.requestHeaders, headers...)
	}
}

// WithCustomFields configures a function that returns additional fields to
// include in the request completion log. The function is called after the
// request is processed, so it has access to any values set during handling.
//
// Example:
//
//	ginmw.LoggingMiddleware(tel,
//	    ginmw.WithCustomFields(func(c *gin.Context) map[string]interface{} {
//	        return map[string]interface{}{
//	            "user_id":    c.GetString("user_id"),
//	            "request_id": c.GetString("request_id"),
//	        }
//	    }),
//	)
func WithCustomFields(fn func(*gin.Context) map[string]interface{}) Option {
	return func(c *config) {
		c.customFields = fn
	}
}
