package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	otelx "github.com/twistingmercury/otelx/src/v2"
	"github.com/twistingmercury/otelx_examples/gingonic/simple/internal/server"
)

func main() {
	ctx := context.Background()

	// Initialize telemetry
	// The OTLP endpoint and TLS settings come from OTEL_EXPORTER_OTLP_ENDPOINT
	// and OTEL_EXPORTER_OTLP_INSECURE.
	tel, err := otelx.Initialize(ctx, otelx.Config{
		ServiceName:    "uuid-api",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		LogLevel:       "debug",
		MetricsAddr:    ":9090",
		Tracing:        true,
	})
	if err != nil {
		log.Fatalf("failed to initialize telemetry: %s", err.Error())
	}
	defer func() {
		if err := tel.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown telemetry: %s", err.Error())
		}
	}()

	tel.Logger.Info().Str("metrics_url", tel.MetricsURL()).Msg("telemetry initialized")

	// Initialize and run server
	engine := gin.New()
	if err := server.Initialize(engine, tel); err != nil {
		log.Fatalf("failed to setup server: %v", err)
	}
	if err := server.Run(ctx, engine, tel); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
