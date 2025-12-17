package otelx_test

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twistingmercury/otelx"
)

func TestWithService(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	opt := otelx.WithService("my-service", "1.2.3", "production")

	err := opt(cfg)
	require.NoError(t, err)

	assert.Equal(t, "my-service", cfg.ServiceName)
	assert.Equal(t, "1.2.3", cfg.ServiceVersion)
	assert.Equal(t, "production", cfg.Environment)
}

func TestWithLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    zerolog.Level
		expected zerolog.Level
	}{
		{"trace", zerolog.TraceLevel, zerolog.TraceLevel},
		{"debug", zerolog.DebugLevel, zerolog.DebugLevel},
		{"info", zerolog.InfoLevel, zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel, zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel, zerolog.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := otelx.NewDefaultConfig()
			opt := otelx.WithLogLevel(tt.level)

			err := opt(cfg)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, cfg.LogLevel)
		})
	}
}

func TestWithLogWriter(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	buf := &bytes.Buffer{}
	opt := otelx.WithLogWriter(buf)

	err := opt(cfg)
	require.NoError(t, err)

	assert.Equal(t, buf, cfg.LogWriter)
}

func TestWithoutLogging(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	assert.True(t, cfg.LoggingEnabled) // Default is enabled

	opt := otelx.WithoutLogging()
	err := opt(cfg)
	require.NoError(t, err)

	assert.False(t, cfg.LoggingEnabled)
}

func TestWithMetrics(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	assert.False(t, cfg.MetricsEnabled) // Default is disabled

	opt := otelx.WithMetrics(8080)
	err := opt(cfg)
	require.NoError(t, err)

	assert.True(t, cfg.MetricsEnabled)
	assert.Equal(t, 8080, cfg.MetricsPort)
}

func TestWithMetricsPath(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	opt := otelx.WithMetricsPath("/custom/path")

	err := opt(cfg)
	require.NoError(t, err)

	assert.Equal(t, "/custom/path", cfg.MetricsPath)
}

func TestWithTracing(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	assert.False(t, cfg.TracingEnabled) // Default is disabled

	opt := otelx.WithTracing()
	err := opt(cfg)
	require.NoError(t, err)

	assert.True(t, cfg.TracingEnabled)
}

func TestWithTraceSampleRate(t *testing.T) {
	tests := []struct {
		name string
		rate float64
	}{
		{"zero", 0.0},
		{"half", 0.5},
		{"full", 1.0},
		{"ten percent", 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := otelx.NewDefaultConfig()
			opt := otelx.WithTraceSampleRate(tt.rate)

			err := opt(cfg)
			require.NoError(t, err)

			assert.Equal(t, tt.rate, cfg.TraceSampleRate)
		})
	}
}

func TestWithTraceExporter(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	exporter := "mock-exporter"
	opt := otelx.WithTraceExporter(exporter)

	err := opt(cfg)
	require.NoError(t, err)

	assert.Equal(t, exporter, cfg.TraceExporter)
}

func TestWithOTLPEndpoint(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	opt := otelx.WithOTLPEndpoint("collector.example.com:4317")

	err := opt(cfg)
	require.NoError(t, err)

	assert.Equal(t, "collector.example.com:4317", cfg.OTLPEndpoint)
}

func TestWithOTLPInsecure(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	assert.False(t, cfg.OTLPInsecure) // Default is secure

	opt := otelx.WithOTLPInsecure()
	err := opt(cfg)
	require.NoError(t, err)

	assert.True(t, cfg.OTLPInsecure)
}

func TestWithAllSignals(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	opt := otelx.WithAllSignals()

	err := opt(cfg)
	require.NoError(t, err)

	assert.True(t, cfg.MetricsEnabled)
	assert.Equal(t, otelx.DefaultMetricsPort, cfg.MetricsPort)
	assert.True(t, cfg.TracingEnabled)
}

func TestWithDevelopmentDefaults(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	opt := otelx.WithDevelopmentDefaults()

	err := opt(cfg)
	require.NoError(t, err)

	assert.Equal(t, zerolog.DebugLevel, cfg.LogLevel)
	assert.True(t, cfg.OTLPInsecure)
	assert.Equal(t, 1.0, cfg.TraceSampleRate)
}

func TestWithProductionDefaults(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	opt := otelx.WithProductionDefaults()

	err := opt(cfg)
	require.NoError(t, err)

	assert.Equal(t, zerolog.InfoLevel, cfg.LogLevel)
	assert.False(t, cfg.OTLPInsecure)
	assert.Equal(t, 0.1, cfg.TraceSampleRate)
}

func TestOptionsChaining(t *testing.T) {
	cfg := otelx.NewDefaultConfig()

	opts := []otelx.Option{
		otelx.WithService("my-service", "1.0.0", "staging"),
		otelx.WithLogLevel(zerolog.DebugLevel),
		otelx.WithMetrics(9090),
		otelx.WithMetricsPath("/prometheus"),
		otelx.WithTracing(),
		otelx.WithTraceSampleRate(0.5),
		otelx.WithOTLPEndpoint("otel-collector:4317"),
		otelx.WithOTLPInsecure(),
	}

	for _, opt := range opts {
		err := opt(cfg)
		require.NoError(t, err)
	}

	// Verify all options were applied
	assert.Equal(t, "my-service", cfg.ServiceName)
	assert.Equal(t, "1.0.0", cfg.ServiceVersion)
	assert.Equal(t, "staging", cfg.Environment)
	assert.Equal(t, zerolog.DebugLevel, cfg.LogLevel)
	assert.True(t, cfg.MetricsEnabled)
	assert.Equal(t, 9090, cfg.MetricsPort)
	assert.Equal(t, "/prometheus", cfg.MetricsPath)
	assert.True(t, cfg.TracingEnabled)
	assert.Equal(t, 0.5, cfg.TraceSampleRate)
	assert.Equal(t, "otel-collector:4317", cfg.OTLPEndpoint)
	assert.True(t, cfg.OTLPInsecure)
}

func TestOptionsOrder(t *testing.T) {
	// Later options should override earlier ones
	cfg := otelx.NewDefaultConfig()

	opts := []otelx.Option{
		otelx.WithProductionDefaults(),         // Sets InfoLevel
		otelx.WithLogLevel(zerolog.DebugLevel), // Override to Debug
	}

	for _, opt := range opts {
		err := opt(cfg)
		require.NoError(t, err)
	}

	assert.Equal(t, zerolog.DebugLevel, cfg.LogLevel)
}
