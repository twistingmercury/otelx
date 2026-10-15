package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	otelx "github.com/twistingmercury/otelx/src/v2"
	"github.com/twistingmercury/otelx_examples/gingonic/simple/internal/server"
)

// shutdownTimeout bounds the telemetry flush when the service stops.
const shutdownTimeout = 10 * time.Second

func main() {
	// SIGINT and SIGTERM (what docker stop and Kubernetes send) cancel ctx, so
	// server.Run returns and the deferred telemetry shutdown below runs.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	defer shutdownTelemetry(tel)

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

// shutdownTelemetry flushes pending telemetry. It uses a fresh context because
// the signal context is already canceled by the time it runs.
func shutdownTelemetry(tel *otelx.Telemetry) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := tel.Shutdown(ctx); err != nil {
		log.Printf("failed to shutdown telemetry: %s", err.Error())
	}
}
