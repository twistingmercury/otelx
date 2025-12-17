package otelx_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twistingmercury/otelx"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := otelx.NewDefaultConfig()

	assert.Equal(t, zerolog.InfoLevel, cfg.LogLevel)
	assert.Equal(t, os.Stdout, cfg.LogWriter)
	assert.True(t, cfg.LoggingEnabled)
	assert.Equal(t, otelx.DefaultMetricsPath, cfg.MetricsPath)
	assert.Equal(t, otelx.DefaultTraceSampleRate, cfg.TraceSampleRate)
	assert.Equal(t, otelx.DefaultOTLPEndpoint, cfg.OTLPEndpoint)
	assert.False(t, cfg.MetricsEnabled)
	assert.False(t, cfg.TracingEnabled)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*otelx.Config)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid minimal config",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
			},
			wantErr: false,
		},
		{
			name: "missing service name",
			configure: func(c *otelx.Config) {
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
			},
			wantErr: true,
			errMsg:  "service name is required",
		},
		{
			name: "missing service version",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.Environment = "test"
			},
			wantErr: true,
			errMsg:  "service version is required",
		},
		{
			name: "missing environment",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
			},
			wantErr: true,
			errMsg:  "environment is required",
		},
		{
			name: "valid config with metrics",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
				c.MetricsEnabled = true
				c.MetricsPort = 9090
				c.MetricsPath = "/metrics"
			},
			wantErr: false,
		},
		{
			name: "invalid metrics port too low",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
				c.MetricsEnabled = true
				c.MetricsPort = 0
			},
			wantErr: true,
			errMsg:  "metrics port must be between 1 and 65535",
		},
		{
			name: "invalid metrics port too high",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
				c.MetricsEnabled = true
				c.MetricsPort = 70000
			},
			wantErr: true,
			errMsg:  "metrics port must be between 1 and 65535",
		},
		{
			name: "empty metrics path",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
				c.MetricsEnabled = true
				c.MetricsPort = 9090
				c.MetricsPath = ""
			},
			wantErr: true,
			errMsg:  "metrics path cannot be empty",
		},
		{
			name: "valid config with tracing",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
				c.TracingEnabled = true
				c.OTLPEndpoint = "localhost:4317"
				c.TraceSampleRate = 0.5
			},
			wantErr: false,
		},
		{
			name: "invalid trace sample rate too low",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
				c.TracingEnabled = true
				c.OTLPEndpoint = "localhost:4317"
				c.TraceSampleRate = -0.1
			},
			wantErr: true,
			errMsg:  "trace sample rate must be between 0.0 and 1.0",
		},
		{
			name: "invalid trace sample rate too high",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
				c.TracingEnabled = true
				c.OTLPEndpoint = "localhost:4317"
				c.TraceSampleRate = 1.5
			},
			wantErr: true,
			errMsg:  "trace sample rate must be between 0.0 and 1.0",
		},
		{
			name: "missing OTLP endpoint with tracing enabled",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
				c.TracingEnabled = true
				c.OTLPEndpoint = ""
			},
			wantErr: true,
			errMsg:  "OTLP endpoint is required when tracing is enabled",
		},
		{
			name: "nil log writer with logging enabled",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
				c.LoggingEnabled = true
				c.LogWriter = nil
			},
			wantErr: true,
			errMsg:  "log writer cannot be nil when logging is enabled",
		},
		{
			name: "nil log writer with logging disabled",
			configure: func(c *otelx.Config) {
				c.ServiceName = "test-service"
				c.ServiceVersion = "1.0.0"
				c.Environment = "test"
				c.LoggingEnabled = false
				c.LogWriter = nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := otelx.NewDefaultConfig()
			tt.configure(cfg)

			err := cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_Clone(t *testing.T) {
	original := otelx.NewDefaultConfig()
	original.ServiceName = "test-service"
	original.ServiceVersion = "1.0.0"
	original.Environment = "test"
	original.LogLevel = zerolog.DebugLevel
	original.MetricsEnabled = true
	original.MetricsPort = 8080
	original.TracingEnabled = true
	original.TraceSampleRate = 0.5

	cloned := original.Clone()

	// Verify all fields are copied
	assert.Equal(t, original.ServiceName, cloned.ServiceName)
	assert.Equal(t, original.ServiceVersion, cloned.ServiceVersion)
	assert.Equal(t, original.Environment, cloned.Environment)
	assert.Equal(t, original.LogLevel, cloned.LogLevel)
	assert.Equal(t, original.LoggingEnabled, cloned.LoggingEnabled)
	assert.Equal(t, original.MetricsEnabled, cloned.MetricsEnabled)
	assert.Equal(t, original.MetricsPort, cloned.MetricsPort)
	assert.Equal(t, original.MetricsPath, cloned.MetricsPath)
	assert.Equal(t, original.TracingEnabled, cloned.TracingEnabled)
	assert.Equal(t, original.TraceSampleRate, cloned.TraceSampleRate)
	assert.Equal(t, original.OTLPEndpoint, cloned.OTLPEndpoint)
	assert.Equal(t, original.OTLPInsecure, cloned.OTLPInsecure)

	// Verify modification of clone doesn't affect original
	cloned.ServiceName = "modified"
	cloned.MetricsPort = 9999
	assert.Equal(t, "test-service", original.ServiceName)
	assert.Equal(t, 8080, original.MetricsPort)
}

func TestConfig_ValidateMultipleErrors(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	cfg.LogWriter = &bytes.Buffer{} // Prevent nil writer error

	err := cfg.Validate()
	require.Error(t, err)

	// Should contain all three missing field errors
	assert.Contains(t, err.Error(), "service name is required")
	assert.Contains(t, err.Error(), "service version is required")
	assert.Contains(t, err.Error(), "environment is required")
}
