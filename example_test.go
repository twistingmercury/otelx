package otelx_test

import (
	"context"
	"log"

	"github.com/twistingmercury/otelx/v2"
	"go.opentelemetry.io/otel"
)

func Example() {
	ctx := context.Background()

	// Exporter endpoint, TLS and sampling come from the standard
	// OTEL_EXPORTER_OTLP_* and OTEL_TRACES_SAMPLER environment variables.
	tel, err := otelx.Initialize(ctx, otelx.Config{
		ServiceName:    "my-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		MetricsAddr:    ":9090",
		Tracing:        true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = tel.Shutdown(ctx) }()

	ctx, span := otel.Tracer("my-service").Start(ctx, "startup")
	defer span.End()

	logger := otelx.LoggerWithSpan(tel.Logger, ctx)
	logger.Info().Msg("service started")
}
