package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/twistingmercury/otelx"
	"github.com/twistingmercury/otelx_examples/gingonic/internal/server"
)

func main() {
	ctx := context.Background()

	// Get OTLP endpoint from env or use default
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	// Initialize telemetry
	tel, err := otelx.Initialize(ctx,
		otelx.WithService("uuid-api", "1.0.0", "development"),
		otelx.WithDevelopmentDefaults(),
		otelx.WithMetrics(9090),
		otelx.WithTracing(),
		otelx.WithOTLPEndpoint(endpoint),
	)
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
