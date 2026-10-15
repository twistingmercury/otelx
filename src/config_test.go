package otelx

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Defaults(t *testing.T) {
	cfg, err := Config{ServiceName: "svc"}.withDefaults()
	require.NoError(t, err)

	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, io.Writer(os.Stdout), cfg.LogWriter)
	assert.Equal(t, "/metrics", cfg.MetricsPath)
	assert.Empty(t, cfg.MetricsAddr, "metrics stay off by default")
	assert.False(t, cfg.Tracing, "tracing stays off by default")
}

func TestConfig_RequiresServiceName(t *testing.T) {
	_, err := Config{}.withDefaults()
	require.ErrorContains(t, err, "ServiceName")
}

func TestConfig_ServiceNameFromEnv(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "from-env")
	_, err := Config{}.withDefaults()
	require.NoError(t, err)
}

// The SDK trims OTEL_SERVICE_NAME and ignores it when nothing is left, so a
// blank value must not satisfy the service name requirement.
func TestConfig_BlankServiceNameFromEnv(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "   ")
	_, err := Config{}.withDefaults()
	require.ErrorContains(t, err, "ServiceName")
}

func TestConfig_BlankServiceNameEnvWithConfigName(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "   ")
	_, err := Config{ServiceName: "svc"}.withDefaults()
	require.NoError(t, err)
}

func TestConfig_InvalidLogLevel(t *testing.T) {
	_, err := Config{ServiceName: "svc", LogLevel: "loud"}.withDefaults()
	require.ErrorContains(t, err, "LogLevel")
}
