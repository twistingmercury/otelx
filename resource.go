package otelx

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// NewResource creates an OpenTelemetry resource with standard semantic conventions.
// The resource includes service identity, host information, process details,
// and SDK metadata.
func NewResource(ctx context.Context, cfg *Config) (*resource.Resource, error) {
	hostname, _ := os.Hostname()

	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
			semconv.HostName(hostname),
			semconv.ProcessPID(os.Getpid()),
			semconv.TelemetrySDKName("otelx"),
			semconv.TelemetrySDKVersion("1.0.0"),
			semconv.TelemetrySDKLanguageGo,
		),
	)
}
