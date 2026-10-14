package otelx

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// NewResource creates an OpenTelemetry resource with standard semantic conventions.
// The resource includes service identity, host information, process details,
// and the OpenTelemetry SDK's own telemetry.sdk.* attributes.
//
// Attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME are applied
// last, so they override the values taken from cfg.
func NewResource(ctx context.Context, cfg *Config) (*resource.Resource, error) {
	hostname, _ := os.Hostname()

	return resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentNameKey.String(cfg.Environment),
			semconv.HostName(hostname),
			semconv.ProcessPID(os.Getpid()),
		),
		resource.WithFromEnv(),
	)
}
