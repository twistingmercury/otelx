package otelx

import (
	"io"

	"github.com/rs/zerolog"
)

// Option is a function that configures otelx initialization.
type Option func(*Config) error

// WithService sets the service identity (required).
// This is the only required option - it sets the service name, version, and environment
// which are used in OpenTelemetry resource attributes and logging context.
func WithService(name, version, environment string) Option {
	return func(c *Config) error {
		c.ServiceName = name
		c.ServiceVersion = version
		c.Environment = environment
		return nil
	}
}

// WithLogLevel sets the minimum log level.
// Logs below this level will not be emitted.
// Default is zerolog.InfoLevel.
func WithLogLevel(level zerolog.Level) Option {
	return func(c *Config) error {
		c.LogLevel = level
		return nil
	}
}

// WithLogWriter sets the output writer for logs.
// Default is os.Stdout.
func WithLogWriter(w io.Writer) Option {
	return func(c *Config) error {
		c.LogWriter = w
		return nil
	}
}

// WithoutLogging disables logging entirely.
// When disabled, the Logger in Telemetry will be a no-op logger.
func WithoutLogging() Option {
	return func(c *Config) error {
		c.LoggingEnabled = false
		return nil
	}
}

// WithMetrics enables Prometheus metrics on the specified port.
// A metrics HTTP server will be started serving on /metrics (or custom path).
func WithMetrics(port int) Option {
	return func(c *Config) error {
		c.MetricsEnabled = true
		c.MetricsPort = port
		return nil
	}
}

// WithMetricsPath sets the HTTP path for the metrics endpoint.
// Default is "/metrics".
func WithMetricsPath(path string) Option {
	return func(c *Config) error {
		c.MetricsPath = path
		return nil
	}
}

// WithTracing enables OTLP tracing.
// Traces will be exported to the configured OTLP endpoint.
func WithTracing() Option {
	return func(c *Config) error {
		c.TracingEnabled = true
		return nil
	}
}

// WithTraceSampleRate sets the trace sampling rate.
// Value must be between 0.0 (no sampling) and 1.0 (sample everything).
// Default is 1.0.
func WithTraceSampleRate(rate float64) Option {
	return func(c *Config) error {
		c.TraceSampleRate = rate
		return nil
	}
}

// WithTraceExporter sets a custom trace exporter.
// When set, the default OTLP exporter will not be created.
// The exporter must implement sdktrace.SpanExporter.
func WithTraceExporter(exporter interface{}) Option {
	return func(c *Config) error {
		c.TraceExporter = exporter
		return nil
	}
}

// WithOTLPEndpoint sets the OTLP collector endpoint for tracing.
// Default is "localhost:4317".
func WithOTLPEndpoint(endpoint string) Option {
	return func(c *Config) error {
		c.OTLPEndpoint = endpoint
		return nil
	}
}

// WithOTLPInsecure disables TLS for OTLP connections.
// Use this for local development or when connecting to a collector without TLS.
func WithOTLPInsecure() Option {
	return func(c *Config) error {
		c.OTLPInsecure = true
		return nil
	}
}

// WithAllSignals enables both metrics (on port 9090) and tracing.
// This is a convenience option equivalent to:
//
//	WithMetrics(9090), WithTracing()
func WithAllSignals() Option {
	return func(c *Config) error {
		c.MetricsEnabled = true
		c.MetricsPort = DefaultMetricsPort
		c.TracingEnabled = true
		return nil
	}
}

// WithDevelopmentDefaults configures otelx for local development:
//   - Debug log level
//   - Pretty-printed logs (console writer)
//   - Insecure OTLP connections
//   - 100% trace sampling
func WithDevelopmentDefaults() Option {
	return func(c *Config) error {
		c.LogLevel = zerolog.DebugLevel
		c.OTLPInsecure = true
		c.TraceSampleRate = 1.0
		return nil
	}
}

// WithProductionDefaults configures otelx for production:
//   - Info log level
//   - JSON logs
//   - TLS enabled (not insecure)
//   - 10% trace sampling
func WithProductionDefaults() Option {
	return func(c *Config) error {
		c.LogLevel = zerolog.InfoLevel
		c.OTLPInsecure = false
		c.TraceSampleRate = 0.1
		return nil
	}
}
