package otelx_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twistingmercury/otelx/src/v2"
	"github.com/twistingmercury/otelx/src/v2/internal/testutil"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

func initTracing(t *testing.T, cfg otelx.Config) (*otelx.Telemetry, *tracetest.InMemoryExporter) {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	cfg.ServiceName = "svc"
	cfg.LogWriter = testutil.NopWriter{}
	cfg.Tracing = true
	cfg.TraceExporter = exporter

	tel, err := otelx.Initialize(context.Background(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tel.Shutdown(context.Background()) })
	return tel, exporter
}

func TestTracing_ExportsSpansWithResource(t *testing.T) {
	ctx := context.Background()
	tel, exporter := initTracing(t, otelx.Config{ServiceVersion: "1.2.3"})

	_, span := tel.TracerProvider.Tracer("test").Start(ctx, "test-span")
	span.End()
	require.NoError(t, tel.TracerProvider.ForceFlush(ctx))

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test-span", spans[0].Name)

	attrs := map[string]string{}
	for _, kv := range spans[0].Resource.Attributes() {
		attrs[string(kv.Key)] = kv.Value.String()
	}
	assert.Equal(t, "svc", attrs["service.name"])
	assert.Equal(t, "1.2.3", attrs["service.version"])
}

// A sampled remote parent must keep its child sampled even when the local
// sampler would drop a root span.
func TestTracing_RespectsParentSamplingDecision(t *testing.T) {
	ctx := context.Background()
	t.Setenv("OTEL_TRACES_SAMPLER", "parentbased_traceidratio")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "0")
	tel, _ := initTracing(t, otelx.Config{})

	sampledParent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01},
		SpanID:     trace.SpanID{0x01},
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	parentCtx := trace.ContextWithRemoteSpanContext(ctx, sampledParent)

	_, child := tel.TracerProvider.Tracer("test").Start(parentCtx, "child")
	assert.True(t, child.SpanContext().IsSampled(), "child of a sampled parent must be sampled")
	child.End()

	_, root := tel.TracerProvider.Tracer("test").Start(ctx, "root")
	assert.False(t, root.SpanContext().IsSampled(), "root span must follow the 0 ratio")
	root.End()
}

func TestTracing_ConfigSamplerOverridesEnvironment(t *testing.T) {
	t.Setenv("OTEL_TRACES_SAMPLER", "always_on")
	tel, _ := initTracing(t, otelx.Config{Sampler: sdktrace.NeverSample()})

	_, span := tel.TracerProvider.Tracer("test").Start(context.Background(), "span")
	assert.False(t, span.SpanContext().IsSampled())
	span.End()
}

// fakeCollector is an OTLP gRPC trace endpoint that counts exported spans.
type fakeCollector struct {
	collectortrace.UnimplementedTraceServiceServer
	spans chan int
}

func (c *fakeCollector) Export(_ context.Context, req *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	n := 0
	for _, rs := range req.ResourceSpans {
		for _, ss := range rs.ScopeSpans {
			n += len(ss.Spans)
		}
	}
	c.spans <- n
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

func TestTracing_DefaultExporterUsesEndpointFromEnvironment(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	collector := &fakeCollector{spans: make(chan int, 1)}
	server := grpc.NewServer()
	collectortrace.RegisterTraceServiceServer(server, collector)
	go func() { _ = server.Serve(ln) }()
	defer server.Stop()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+ln.Addr().String())

	ctx := context.Background()
	tel, err := otelx.Initialize(ctx, otelx.Config{ServiceName: "svc", LogWriter: testutil.NopWriter{}, Tracing: true})
	require.NoError(t, err)

	_, span := tel.TracerProvider.Tracer("test").Start(ctx, "span")
	span.End()
	require.NoError(t, tel.Shutdown(ctx))

	select {
	case n := <-collector.spans:
		assert.Equal(t, 1, n)
	case <-time.After(5 * time.Second):
		t.Fatal("span was not exported to OTEL_EXPORTER_OTLP_ENDPOINT")
	}
}
