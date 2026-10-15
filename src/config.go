package otelx

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const defaultMetricsPath = "/metrics"

// Config configures Initialize. The zero value is valid apart from
// ServiceName, which is required unless OTEL_SERVICE_NAME is set.
//
// Exporter endpoints, headers, TLS and sampling are not set here: they come
// from the standard OTEL_* environment variables, such as
// OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_INSECURE, OTEL_TRACES_SAMPLER
// and OTEL_RESOURCE_ATTRIBUTES.
type Config struct {
	// ServiceName, ServiceVersion and Environment become the service.name,
	// service.version and deployment.environment.name resource attributes.
	// OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES override them.
	ServiceName    string
	ServiceVersion string
	Environment    string

	// LogLevel is a zerolog level name such as "debug", "info" or "disabled".
	// The default is "info".
	LogLevel string

	// LogWriter receives JSON log lines. The default is os.Stdout.
	LogWriter io.Writer

	// MetricsAddr is the address the Prometheus metrics server listens on,
	// such as ":9090". Metrics are disabled when it is empty.
	MetricsAddr string

	// MetricsPath is the HTTP path metrics are served on. It must start with "/"
	// when MetricsAddr is set. The default is "/metrics".
	MetricsPath string

	// Tracing enables the trace provider.
	Tracing bool

	// TraceExporter replaces the default OTLP gRPC exporter when set.
	TraceExporter sdktrace.SpanExporter

	// Sampler replaces the sampler from OTEL_TRACES_SAMPLER when set. Without
	// either, spans are sampled with ParentBased(AlwaysSample).
	Sampler sdktrace.Sampler
}

// withDefaults returns a copy of c with defaults filled in, or an error if c
// is invalid.
func (c Config) withDefaults() (Config, error) {
	var errs []error

	if c.ServiceName == "" && os.Getenv("OTEL_SERVICE_NAME") == "" {
		errs = append(errs, errors.New("ServiceName is required unless OTEL_SERVICE_NAME is set"))
	}

	if c.LogLevel == "" {
		c.LogLevel = zerolog.InfoLevel.String()
	}
	if _, err := zerolog.ParseLevel(c.LogLevel); err != nil {
		errs = append(errs, fmt.Errorf("invalid LogLevel %q: %w", c.LogLevel, err))
	}
	if c.LogWriter == nil {
		c.LogWriter = os.Stdout
	}

	if c.MetricsPath == "" {
		c.MetricsPath = defaultMetricsPath
	}

	// net/http panics on a pattern with no leading slash.
	if c.MetricsAddr != "" && !strings.HasPrefix(c.MetricsPath, "/") {
		errs = append(errs, fmt.Errorf("invalid MetricsPath %q: must start with \"/\"", c.MetricsPath))
	}

	return c, errors.Join(errs...)
}
