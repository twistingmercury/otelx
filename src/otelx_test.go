package otelx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otelx "github.com/twistingmercury/otelx/src/v2"
	"github.com/twistingmercury/otelx/src/v2/internal/testutil"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestInitialize_RequiresServiceName(t *testing.T) {
	tel, err := otelx.Initialize(context.Background(), otelx.Config{})
	require.ErrorContains(t, err, "ServiceName")
	assert.Nil(t, tel)
}

func TestInitialize_LoggingOnlyByDefault(t *testing.T) {
	ctx := context.Background()
	tel, err := otelx.Initialize(ctx, otelx.Config{ServiceName: "svc", LogWriter: testutil.NopWriter{}})
	require.NoError(t, err)
	defer func() { _ = tel.Shutdown(ctx) }()

	assert.Nil(t, tel.TracerProvider)
	assert.Nil(t, tel.MeterProvider)
	assert.Empty(t, tel.MetricsURL())
}

func TestInitialize_AllSignals(t *testing.T) {
	ctx := context.Background()
	tel, err := otelx.Initialize(ctx, otelx.Config{
		ServiceName:   "svc",
		LogWriter:     testutil.NopWriter{},
		MetricsAddr:   "127.0.0.1:0",
		Tracing:       true,
		TraceExporter: tracetest.NewInMemoryExporter(),
	})
	require.NoError(t, err)
	defer func() { _ = tel.Shutdown(ctx) }()

	require.NotNil(t, tel.TracerProvider)
	require.NotNil(t, tel.MeterProvider)
	assert.Same(t, tel.TracerProvider, otel.GetTracerProvider())
	assert.Same(t, tel.MeterProvider, otel.GetMeterProvider())
}

type failingExporter struct{ err error }

func (failingExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error { return nil }
func (e failingExporter) Shutdown(context.Context) error                           { return e.err }

func TestTelemetry_ShutdownWrapsErrors(t *testing.T) {
	ctx := context.Background()
	errExporter := errors.New("exporter shutdown failed")

	tel, err := otelx.Initialize(ctx, otelx.Config{
		ServiceName:   "svc",
		LogWriter:     testutil.NopWriter{},
		Tracing:       true,
		TraceExporter: failingExporter{err: errExporter},
	})
	require.NoError(t, err)

	err = tel.Shutdown(ctx)
	assert.ErrorIs(t, err, errExporter, "Shutdown must keep the underlying error for errors.Is")
}
