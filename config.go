package otelx

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

// Default configuration values.
const (
	DefaultMetricsPath     = "/metrics"
	DefaultMetricsPort     = 9090
	DefaultTraceSampleRate = 1.0
	DefaultOTLPEndpoint    = "localhost:4317"
)

// Config holds the configuration for initializing telemetry components.
// It is populated by applying Option functions to Initialize.
type Config struct {
	// Service identity (required)
	ServiceName    string
	ServiceVersion string
	Environment    string

	// Logging configuration
	LogLevel       zerolog.Level
	LogWriter      io.Writer
	LoggingEnabled bool

	// Metrics configuration
	MetricsEnabled bool
	MetricsPort    int
	MetricsPath    string

	// Tracing configuration
	TracingEnabled  bool
	TraceSampleRate float64
	TraceExporter   interface{} // Custom trace exporter, if provided
	OTLPEndpoint    string
	OTLPInsecure    bool
}

// NewDefaultConfig creates a new Config with default values.
func NewDefaultConfig() *Config {
	return &Config{
		LogLevel:        zerolog.InfoLevel,
		LogWriter:       os.Stdout,
		LoggingEnabled:  true,
		MetricsPath:     DefaultMetricsPath,
		TraceSampleRate: DefaultTraceSampleRate,
		OTLPEndpoint:    DefaultOTLPEndpoint,
	}
}

// Validate checks that the configuration is valid and returns an error
// if any required fields are missing or if values are out of range.
func (c *Config) Validate() error {
	var errs []error

	// Service identity is required
	if c.ServiceName == "" {
		errs = append(errs, errors.New("service name is required: use WithService option"))
	}
	if c.ServiceVersion == "" {
		errs = append(errs, errors.New("service version is required: use WithService option"))
	}
	if c.Environment == "" {
		errs = append(errs, errors.New("environment is required: use WithService option"))
	}

	// Validate metrics configuration
	if c.MetricsEnabled {
		if c.MetricsPort < 1 || c.MetricsPort > 65535 {
			errs = append(errs, fmt.Errorf("metrics port must be between 1 and 65535, got %d", c.MetricsPort))
		}
		if c.MetricsPath == "" {
			errs = append(errs, errors.New("metrics path cannot be empty"))
		}
	}

	// Validate tracing configuration
	if c.TracingEnabled {
		if c.TraceSampleRate < 0.0 || c.TraceSampleRate > 1.0 {
			errs = append(errs, fmt.Errorf("trace sample rate must be between 0.0 and 1.0, got %f", c.TraceSampleRate))
		}
		if c.OTLPEndpoint == "" && c.TraceExporter == nil {
			errs = append(errs, errors.New("OTLP endpoint is required when tracing is enabled: use WithOTLPEndpoint option"))
		}
	}

	// Validate logging configuration
	if c.LoggingEnabled && c.LogWriter == nil {
		errs = append(errs, errors.New("log writer cannot be nil when logging is enabled"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() *Config {
	return &Config{
		ServiceName:     c.ServiceName,
		ServiceVersion:  c.ServiceVersion,
		Environment:     c.Environment,
		LogLevel:        c.LogLevel,
		LogWriter:       c.LogWriter,
		LoggingEnabled:  c.LoggingEnabled,
		MetricsEnabled:  c.MetricsEnabled,
		MetricsPort:     c.MetricsPort,
		MetricsPath:     c.MetricsPath,
		TracingEnabled:  c.TracingEnabled,
		TraceSampleRate: c.TraceSampleRate,
		TraceExporter:   c.TraceExporter,
		OTLPEndpoint:    c.OTLPEndpoint,
		OTLPInsecure:    c.OTLPInsecure,
	}
}
