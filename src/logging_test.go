package otelx_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otelx "github.com/twistingmercury/otelx/src/v2"
	"github.com/twistingmercury/otelx/src/v2/internal/testutil"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func initLogging(t *testing.T, cfg otelx.Config) (*otelx.Telemetry, *testutil.LogCapture) {
	t.Helper()
	capture := testutil.NewLogCapture()
	if cfg.ServiceName == "" {
		cfg.ServiceName = "svc"
	}
	cfg.LogWriter = capture
	tel, err := otelx.Initialize(context.Background(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tel.Shutdown(context.Background()) })
	return tel, capture
}

func TestLogging_JSONWithServiceIdentity(t *testing.T) {
	tel, capture := initLogging(t, otelx.Config{ServiceVersion: "1.2.3", Environment: "test"})
	tel.Logger.Info().Str("k", "v").Msg("hello")

	logs, err := capture.ParseJSONLogs()
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, "hello", logs[0]["message"])
	assert.Equal(t, "info", logs[0]["level"])
	assert.Equal(t, "v", logs[0]["k"])
	assert.Equal(t, "svc", logs[0]["service"])
	assert.Equal(t, "1.2.3", logs[0]["version"])
	assert.Equal(t, "test", logs[0]["environment"])
	assert.Contains(t, logs[0], "time")
}

func TestLogging_ServiceNameFromEnvironment(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "svc-from-env")
	tel, capture := initLogging(t, otelx.Config{})
	tel.Logger.Info().Msg("hello")
	assert.True(t, capture.Contains(`"service":"svc-from-env"`))
}

func TestLogging_Disabled(t *testing.T) {
	tel, capture := initLogging(t, otelx.Config{LogLevel: "disabled"})
	tel.Logger.Error().Msg("dropped")
	assert.Empty(t, capture.String())
}

func TestInitialize_DoesNotChangeZerologGlobalLevel(t *testing.T) {
	before := zerolog.GlobalLevel()
	tel, capture := initLogging(t, otelx.Config{LogLevel: "error"})

	assert.Equal(t, before, zerolog.GlobalLevel())

	tel.Logger.Info().Msg("filtered")
	tel.Logger.Error().Msg("kept")
	assert.False(t, capture.Contains("filtered"))
	assert.True(t, capture.Contains("kept"))
}

func TestLoggerWithSpan(t *testing.T) {
	ctx := context.Background()
	tel, capture := initLogging(t, otelx.Config{Tracing: true, TraceExporter: tracetest.NewInMemoryExporter()})

	noSpan := otelx.LoggerWithSpan(tel.Logger, ctx)
	noSpan.Info().Msg("no span")

	ctx, span := tel.TracerProvider.Tracer("test").Start(ctx, "span")
	defer span.End()
	inSpan := otelx.LoggerWithSpan(tel.Logger, ctx)
	inSpan.Info().Msg("in span")

	logs, err := capture.ParseJSONLogs()
	require.NoError(t, err)
	require.Len(t, logs, 2)
	assert.NotContains(t, logs[0], "trace_id")
	assert.Equal(t, span.SpanContext().TraceID().String(), logs[1]["trace_id"])
	assert.Equal(t, span.SpanContext().SpanID().String(), logs[1]["span_id"])
}
