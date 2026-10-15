package otelx

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// newResource builds the resource shared by traces and metrics. Attributes
// from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME are applied last, so
// they override the values in cfg.
func newResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	hostname, _ := os.Hostname()

	attrs := []attribute.KeyValue{
		semconv.HostName(hostname),
		semconv.ProcessPID(os.Getpid()),
	}
	if cfg.ServiceName != "" {
		attrs = append(attrs, semconv.ServiceName(cfg.ServiceName))
	}
	if cfg.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.ServiceVersion))
	}
	if cfg.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironmentNameKey.String(cfg.Environment))
	}

	return resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(attrs...),
		resource.WithFromEnv(),
	)
}
