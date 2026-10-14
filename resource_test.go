package otelx

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk"
)

func resourceAttrs(t *testing.T, cfg Config) map[string]string {
	t.Helper()
	res, err := newResource(context.Background(), cfg)
	require.NoError(t, err)
	attrs := make(map[string]string)
	for _, kv := range res.Attributes() {
		attrs[string(kv.Key)] = kv.Value.String()
	}
	return attrs
}

func TestNewResource(t *testing.T) {
	attrs := resourceAttrs(t, Config{ServiceName: "svc", ServiceVersion: "1.2.3", Environment: "staging"})

	assert.Equal(t, "svc", attrs["service.name"])
	assert.Equal(t, "1.2.3", attrs["service.version"])
	assert.Equal(t, "staging", attrs["deployment.environment.name"])
	assert.NotContains(t, attrs, "deployment.environment")
	assert.Equal(t, "opentelemetry", attrs["telemetry.sdk.name"])
	assert.Equal(t, sdk.Version(), attrs["telemetry.sdk.version"])
	assert.Equal(t, "go", attrs["telemetry.sdk.language"])
	assert.NotEmpty(t, attrs["process.pid"])
	if hostname, _ := os.Hostname(); hostname != "" {
		assert.Equal(t, hostname, attrs["host.name"])
	}
}

func TestNewResource_OmitsEmptyFields(t *testing.T) {
	attrs := resourceAttrs(t, Config{ServiceName: "svc"})
	assert.NotContains(t, attrs, "service.version")
	assert.NotContains(t, attrs, "deployment.environment.name")
}

func TestNewResource_EnvironmentOverridesConfig(t *testing.T) {
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "team=payments,deployment.environment.name=from-env")
	t.Setenv("OTEL_SERVICE_NAME", "svc-from-env")

	attrs := resourceAttrs(t, Config{ServiceName: "svc", Environment: "from-config"})
	assert.Equal(t, "payments", attrs["team"])
	assert.Equal(t, "from-env", attrs["deployment.environment.name"])
	assert.Equal(t, "svc-from-env", attrs["service.name"])
}
