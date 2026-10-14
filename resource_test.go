package otelx_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twistingmercury/otelx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk"
)

func TestNewResource(t *testing.T) {
	ctx := context.Background()
	cfg := otelx.NewDefaultConfig()
	cfg.ServiceName = "test-service"
	cfg.ServiceVersion = "1.2.3"
	cfg.Environment = "testing"

	res, err := otelx.NewResource(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// Convert resource to attribute map for easier testing
	attrs := res.Attributes()
	attrMap := make(map[attribute.Key]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.AsString()
	}

	// Check service attributes
	assert.Equal(t, "test-service", attrMap["service.name"])
	assert.Equal(t, "1.2.3", attrMap["service.version"])
	assert.Equal(t, "testing", attrMap["deployment.environment.name"])

	assert.Equal(t, "opentelemetry", attrMap["telemetry.sdk.name"])
	assert.Equal(t, sdk.Version(), attrMap["telemetry.sdk.version"])
	assert.Equal(t, "go", attrMap["telemetry.sdk.language"])

	// Check host attributes (hostname may vary)
	hostname, _ := os.Hostname()
	if hostname != "" {
		assert.Equal(t, hostname, attrMap["host.name"])
	}
}

func TestNewResource_DifferentEnvironments(t *testing.T) {
	tests := []struct {
		name        string
		environment string
	}{
		{"development", "development"},
		{"staging", "staging"},
		{"production", "production"},
		{"custom", "my-custom-env"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg := otelx.NewDefaultConfig()
			cfg.ServiceName = "test-service"
			cfg.ServiceVersion = "1.0.0"
			cfg.Environment = tt.environment

			res, err := otelx.NewResource(ctx, cfg)
			require.NoError(t, err)

			attrs := res.Attributes()
			attrMap := make(map[attribute.Key]string)
			for _, attr := range attrs {
				attrMap[attr.Key] = attr.Value.AsString()
			}

			assert.Equal(t, tt.environment, attrMap["deployment.environment.name"])
		})
	}
}

func TestNewResource_AttributeValues(t *testing.T) {
	ctx := context.Background()
	cfg := otelx.NewDefaultConfig()
	cfg.ServiceName = "my-app"
	cfg.ServiceVersion = "2.0.0"
	cfg.Environment = "prod"

	res, err := otelx.NewResource(ctx, cfg)
	require.NoError(t, err)

	// Get attributes as a slice
	attrs := res.Attributes()
	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value.AsString()
	}

	assert.Equal(t, "my-app", attrMap["service.name"])
	assert.Equal(t, "2.0.0", attrMap["service.version"])
	assert.Equal(t, "prod", attrMap["deployment.environment.name"])
	assert.Equal(t, "opentelemetry", attrMap["telemetry.sdk.name"])
	assert.Equal(t, sdk.Version(), attrMap["telemetry.sdk.version"])
	assert.Equal(t, "go", attrMap["telemetry.sdk.language"])
}

func TestNewResource_ProcessPID(t *testing.T) {
	ctx := context.Background()
	cfg := otelx.NewDefaultConfig()
	cfg.ServiceName = "test-service"
	cfg.ServiceVersion = "1.0.0"
	cfg.Environment = "test"

	res, err := otelx.NewResource(ctx, cfg)
	require.NoError(t, err)

	attrs := res.Attributes()
	var foundPID bool
	expectedPID := os.Getpid()

	for _, attr := range attrs {
		if string(attr.Key) == "process.pid" {
			foundPID = true
			assert.Equal(t, int64(expectedPID), attr.Value.AsInt64())
			break
		}
	}

	assert.True(t, foundPID, "process.pid attribute should be present")
}

func resourceAttrs(t *testing.T, cfg *otelx.Config) map[string]string {
	t.Helper()
	res, err := otelx.NewResource(context.Background(), cfg)
	require.NoError(t, err)
	attrs := make(map[string]string)
	for _, kv := range res.Attributes() {
		attrs[string(kv.Key)] = kv.Value.Emit()
	}
	return attrs
}

func TestNewResource_UsesSDKTelemetryAttributes(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	cfg.ServiceName, cfg.ServiceVersion, cfg.Environment = "svc", "1.0.0", "test"

	attrs := resourceAttrs(t, cfg)
	assert.Equal(t, "opentelemetry", attrs["telemetry.sdk.name"])
	assert.Equal(t, sdk.Version(), attrs["telemetry.sdk.version"])
	assert.Equal(t, "go", attrs["telemetry.sdk.language"])
}

func TestNewResource_UsesDeploymentEnvironmentName(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	cfg.ServiceName, cfg.ServiceVersion, cfg.Environment = "svc", "1.0.0", "staging"

	attrs := resourceAttrs(t, cfg)
	assert.Equal(t, "staging", attrs["deployment.environment.name"])
	assert.NotContains(t, attrs, "deployment.environment")
}

func TestNewResource_ReadsOTELResourceAttributes(t *testing.T) {
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "team=payments,deployment.environment.name=from-env")
	cfg := otelx.NewDefaultConfig()
	cfg.ServiceName, cfg.ServiceVersion, cfg.Environment = "svc", "1.0.0", "from-config"

	attrs := resourceAttrs(t, cfg)
	assert.Equal(t, "payments", attrs["team"])
	assert.Equal(t, "from-env", attrs["deployment.environment.name"], "environment variables override config")
}
