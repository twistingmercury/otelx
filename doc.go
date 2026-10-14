// Package otelx sets up OpenTelemetry tracing, Prometheus metrics and a
// trace-aware zerolog logger for a service with one call.
//
// # Quick Start
//
//	tel, err := otelx.Initialize(ctx, otelx.Config{
//	    ServiceName:    "my-service",
//	    ServiceVersion: "1.0.0",
//	    Environment:    "production",
//	    MetricsAddr:    ":9090",
//	    Tracing:        true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer tel.Shutdown(ctx)
//
// Create spans and instruments with otel.Tracer and otel.Meter, and add the
// current span to log lines with LoggerWithSpan.
//
// # Signals
//
//   - Logging: always on. JSON lines from zerolog, tagged with the service
//     identity. otelx does not send logs through the OpenTelemetry logs
//     signal; LoggerWithSpan adds trace_id and span_id fields so logs can be
//     joined to traces.
//   - Metrics: on when Config.MetricsAddr is set. Served in Prometheus format
//     from Config.MetricsPath, together with Go runtime and process metrics.
//   - Tracing: on when Config.Tracing is true. Spans go to the OTLP gRPC
//     exporter, or to Config.TraceExporter when set.
//
// # Environment Variables
//
// otelx leaves standard OpenTelemetry settings to the environment:
//
//   - OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_TRACES_ENDPOINT,
//     OTEL_EXPORTER_OTLP_HEADERS, OTEL_EXPORTER_OTLP_INSECURE and the other
//     OTEL_EXPORTER_OTLP_* variables configure the default exporter.
//   - OTEL_TRACES_SAMPLER and OTEL_TRACES_SAMPLER_ARG choose the sampler
//     unless Config.Sampler is set. The default is ParentBased(AlwaysSample).
//   - OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES add resource attributes
//     and override the values in Config.
//
// # Resource Attributes
//
// Traces and metrics carry service.name, service.version,
// deployment.environment.name, host.name, process.pid and the SDK's
// telemetry.sdk.* attributes, plus anything from the environment.
//
// # Global State
//
// Initialize sets these process-wide globals:
//
//   - otel.SetTracerProvider, when tracing is enabled
//   - otel.SetTextMapPropagator (W3C Trace Context and Baggage), when tracing is enabled
//   - otel.SetMeterProvider, when metrics are enabled
//
// It does not change zerolog's global level or use the default Prometheus
// registry: each Telemetry has its own registry and metrics server. Calling
// Initialize more than once is safe, but the globals above point at the most
// recent call's providers.
//
// # Shutdown
//
// Shutdown flushes and stops the tracer provider, the meter provider and the
// metrics server, and returns every failure joined with errors.Join.
package otelx
