package otelx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twistingmercury/otelx"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTracing_Disabled(t *testing.T) {
	ctx := context.Background()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
		// No WithTracing - tracing disabled
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer func() {
		_ = tel.Shutdown(ctx)
	}()
	assert.Nil(t, tel.TracerProvider)
}

func TestTracing_WithCustomExporter(t *testing.T) {
	ctx := context.Background()

	// Create an in-memory exporter for testing
	exporter := tracetest.NewInMemoryExporter()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
		otelx.WithTracing(),
		otelx.WithTraceExporter(exporter),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer func() {
		_ = tel.Shutdown(ctx)
	}()

	require.NotNil(t, tel.TracerProvider)

	// Create a span
	tracer := tel.TracerProvider.Tracer("test-tracer")
	_, span := tracer.Start(ctx, "test-span")
	span.End()

	// Force flush to ensure span is exported
	err = tel.TracerProvider.ForceFlush(ctx)
	require.NoError(t, err)

	// Verify the span was exported
	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test-span", spans[0].Name)
}

func TestTracing_SampleRates(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate float64
	}{
		{"never sample", 0.0},
		{"always sample", 1.0},
		{"50% sample", 0.5},
		{"10% sample", 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			exporter := tracetest.NewInMemoryExporter()

			tel, err := otelx.Initialize(ctx,
				otelx.WithService("test-service", "1.0.0", "test"),
				otelx.WithoutLogging(),
				otelx.WithTracing(),
				otelx.WithTraceExporter(exporter),
				otelx.WithTraceSampleRate(tt.sampleRate),
			)
			require.NoError(t, err)
			require.NotNil(t, tel)
			defer func() {
				_ = tel.Shutdown(ctx)
			}()

			require.NotNil(t, tel.TracerProvider)
		})
	}
}

func TestTracing_GlobalProvider(t *testing.T) {
	ctx := context.Background()
	exporter := tracetest.NewInMemoryExporter()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
		otelx.WithTracing(),
		otelx.WithTraceExporter(exporter),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer func() {
		_ = tel.Shutdown(ctx)
	}()

	// The global tracer provider should be set
	globalTP := otel.GetTracerProvider()
	require.NotNil(t, globalTP)

	// Should be able to create spans using the global provider
	tracer := globalTP.Tracer("global-tracer")
	_, span := tracer.Start(ctx, "global-span")
	span.End()
}

func TestTracing_SpanAttributes(t *testing.T) {
	ctx := context.Background()
	exporter := tracetest.NewInMemoryExporter()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
		otelx.WithTracing(),
		otelx.WithTraceExporter(exporter),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer func() {
		_ = tel.Shutdown(ctx)
	}()

	// Create a span with attributes
	tracer := tel.TracerProvider.Tracer("test-tracer")
	_, span := tracer.Start(ctx, "attributed-span")
	span.SetName("renamed-span")
	span.End()

	// Force flush
	err = tel.TracerProvider.ForceFlush(ctx)
	require.NoError(t, err)

	// Verify
	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "renamed-span", spans[0].Name)
}

func TestTracing_NestedSpans(t *testing.T) {
	ctx := context.Background()
	exporter := tracetest.NewInMemoryExporter()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
		otelx.WithTracing(),
		otelx.WithTraceExporter(exporter),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer func() {
		_ = tel.Shutdown(ctx)
	}()

	tracer := tel.TracerProvider.Tracer("test-tracer")

	// Create parent span
	parentCtx, parentSpan := tracer.Start(ctx, "parent-span")

	// Create child span
	_, childSpan := tracer.Start(parentCtx, "child-span")
	childSpan.End()

	parentSpan.End()

	// Force flush
	err = tel.TracerProvider.ForceFlush(ctx)
	require.NoError(t, err)

	// Verify both spans
	spans := exporter.GetSpans()
	require.Len(t, spans, 2)

	// Find spans by name
	var parent, child *tracetest.SpanStub
	for i := range spans {
		switch spans[i].Name {
		case "parent-span":
			parent = &spans[i]
		case "child-span":
			child = &spans[i]
		}
	}

	require.NotNil(t, parent)
	require.NotNil(t, child)

	// Child should have parent's trace ID
	assert.Equal(t, parent.SpanContext.TraceID(), child.SpanContext.TraceID())
	// Child should reference parent
	assert.Equal(t, parent.SpanContext.SpanID(), child.Parent.SpanID())
}

// MockSpanExporter is a mock implementation for testing invalid exporter types
type MockSpanExporter struct{}

func (m *MockSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (m *MockSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

func TestTracing_InvalidExporterType(t *testing.T) {
	ctx := context.Background()

	// Pass something that's not a SpanExporter
	_, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
		otelx.WithTracing(),
		otelx.WithTraceExporter("not an exporter"),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SpanExporter")
}
