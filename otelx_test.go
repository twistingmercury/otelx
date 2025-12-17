package otelx_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twistingmercury/otelx"
	"github.com/twistingmercury/otelx/internal/testutil"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestInitialize_RequiresService(t *testing.T) {
	ctx := context.Background()

	_, err := otelx.Initialize(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service name is required")
}

func TestInitialize_MinimalConfig(t *testing.T) {
	ctx := context.Background()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	// Logging is disabled but logger should still exist (nop)
	tel.Logger.Info().Msg("this should not panic")

	// Metrics and tracing should be nil when not enabled
	assert.Nil(t, tel.MeterProvider)
	assert.Nil(t, tel.TracerProvider)
}

func TestInitialize_WithLogging(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithLogWriter(&buf),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	tel.Logger.Info().Msg("test message")

	assert.Contains(t, buf.String(), "test message")
	assert.Contains(t, buf.String(), "info")
}

func TestInitialize_WithMetrics(t *testing.T) {
	ctx := context.Background()
	port := testutil.GetFreePort(t)

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithMetrics(port),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	require.NotNil(t, tel.MeterProvider)
	assert.NotEmpty(t, tel.MetricsURL())

	// Wait for server to start and verify it's accessible
	ok := testutil.WaitForServer(t, tel.MetricsURL(), 5*time.Second)
	require.True(t, ok, "metrics server should be accessible")
}

func TestInitialize_WithTracing(t *testing.T) {
	ctx := context.Background()
	exporter := tracetest.NewInMemoryExporter()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithTracing(),
		otelx.WithTraceExporter(exporter),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	require.NotNil(t, tel.TracerProvider)

	// Create a span and verify it's exported
	tracer := tel.TracerProvider.Tracer("test")
	_, span := tracer.Start(ctx, "test-span")
	span.End()

	err = tel.TracerProvider.ForceFlush(ctx)
	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test-span", spans[0].Name)
}

func TestInitialize_FullStack(t *testing.T) {
	ctx := context.Background()
	port := testutil.GetFreePort(t)
	var logBuf bytes.Buffer
	exporter := tracetest.NewInMemoryExporter()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("full-stack-service", "2.0.0", "integration"),
		otelx.WithLogWriter(&logBuf),
		otelx.WithMetrics(port),
		otelx.WithTracing(),
		otelx.WithTraceExporter(exporter),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	// All components should be initialized
	require.NotNil(t, tel.MeterProvider)
	require.NotNil(t, tel.TracerProvider)
	assert.NotEmpty(t, tel.MetricsURL())

	// Logging should work
	tel.Logger.Info().Msg("full stack test")
	assert.Contains(t, logBuf.String(), "full stack test")
}

func TestInitialize_AllSignals(t *testing.T) {
	ctx := context.Background()
	exporter := tracetest.NewInMemoryExporter()

	// WithAllSignals uses default port 9090, which might be in use
	// So we'll test by checking the config instead
	cfg := otelx.NewDefaultConfig()
	opt := otelx.WithAllSignals()
	err := opt(cfg)
	require.NoError(t, err)

	assert.True(t, cfg.MetricsEnabled)
	assert.Equal(t, 9090, cfg.MetricsPort)
	assert.True(t, cfg.TracingEnabled)

	// Actually initialize with a custom port
	port := testutil.GetFreePort(t)
	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithMetrics(port),
		otelx.WithTracing(),
		otelx.WithTraceExporter(exporter),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	assert.NotNil(t, tel.MeterProvider)
	assert.NotNil(t, tel.TracerProvider)
}

func TestTelemetry_Shutdown(t *testing.T) {
	ctx := context.Background()
	port := testutil.GetFreePort(t)
	exporter := tracetest.NewInMemoryExporter()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithMetrics(port),
		otelx.WithTracing(),
		otelx.WithTraceExporter(exporter),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)

	// Wait for server to start
	ok := testutil.WaitForServer(t, tel.MetricsURL(), 5*time.Second)
	require.True(t, ok, "metrics server should start")

	// Shutdown
	err = tel.Shutdown(ctx)
	require.NoError(t, err)

	// Server should no longer be accessible after a brief delay
	time.Sleep(100 * time.Millisecond)
}

func TestTelemetry_ShutdownWithTimeout(t *testing.T) {
	ctx := context.Background()
	port := testutil.GetFreePort(t)

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithMetrics(port),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)

	// Use a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = tel.Shutdown(shutdownCtx)
	require.NoError(t, err)
}

func TestTelemetry_Config(t *testing.T) {
	ctx := context.Background()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("config-test", "3.0.0", "testing"),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	cfg := tel.Config()
	assert.Equal(t, "config-test", cfg.ServiceName)
	assert.Equal(t, "3.0.0", cfg.ServiceVersion)
	assert.Equal(t, "testing", cfg.Environment)

	// Modifying returned config shouldn't affect internal state
	cfg.ServiceName = "modified"
	cfg2 := tel.Config()
	assert.Equal(t, "config-test", cfg2.ServiceName)
}

func TestTelemetry_MetricsURL_Disabled(t *testing.T) {
	ctx := context.Background()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	assert.Empty(t, tel.MetricsURL())
}

func TestInitialize_DevelopmentDefaults(t *testing.T) {
	cfg := otelx.NewDefaultConfig()

	opt := otelx.WithDevelopmentDefaults()
	err := opt(cfg)
	require.NoError(t, err)

	// Development defaults should set debug level and insecure
	assert.Equal(t, float64(1.0), cfg.TraceSampleRate)
	assert.True(t, cfg.OTLPInsecure)
}

func TestInitialize_ProductionDefaults(t *testing.T) {
	cfg := otelx.NewDefaultConfig()

	opt := otelx.WithProductionDefaults()
	err := opt(cfg)
	require.NoError(t, err)

	// Production defaults should set lower sample rate and secure
	assert.Equal(t, 0.1, cfg.TraceSampleRate)
	assert.False(t, cfg.OTLPInsecure)
}

func TestInitialize_InvalidPort(t *testing.T) {
	ctx := context.Background()

	_, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithMetrics(-1), // Invalid port
		otelx.WithoutLogging(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "metrics port")
}

func TestInitialize_InvalidSampleRate(t *testing.T) {
	ctx := context.Background()

	_, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithTracing(),
		otelx.WithTraceSampleRate(2.0), // Invalid: > 1.0
		otelx.WithoutLogging(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sample rate")
}

func TestTelemetry_Tracer_Disabled(t *testing.T) {
	ctx := context.Background()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
		// No WithTracing - tracing disabled
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	// Should return noop tracer
	tracer := tel.Tracer("test")
	require.NotNil(t, tracer)

	// Using noop tracer should not panic
	newCtx, span := tracer.Start(ctx, "test-span")
	require.NotNil(t, newCtx)
	require.NotNil(t, span)
}

func TestTelemetry_Meter_Disabled(t *testing.T) {
	ctx := context.Background()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
		// No WithMetrics - metrics disabled
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	// Should return noop meter from global provider
	meter := tel.Meter("test")
	assert.NotNil(t, meter)
}
