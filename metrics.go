package otelx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// MetricsServer wraps an HTTP server that serves Prometheus metrics.
type MetricsServer struct {
	server *http.Server
	port   int
	path   string
}

// setupMetrics configures the Prometheus metrics exporter and returns a MeterProvider.
// It also starts an HTTP server to serve the /metrics endpoint.
func setupMetrics(ctx context.Context, cfg *Config, res *resource.Resource) (*metric.MeterProvider, *MetricsServer, error) {
	if !cfg.MetricsEnabled {
		return nil, nil, nil
	}

	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create MeterProvider with the exporter
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(exporter),
	)

	// Create HTTP server for metrics
	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, promhttp.Handler())

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	metricsServer := &MetricsServer{
		server: server,
		port:   cfg.MetricsPort,
		path:   cfg.MetricsPath,
	}

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Log error but don't panic - the server might be intentionally stopped
			fmt.Printf("metrics server error: %v\n", err)
		}
	}()

	return mp, metricsServer, nil
}

// Shutdown gracefully shuts down the metrics server.
func (s *MetricsServer) Shutdown(ctx context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// Port returns the port the metrics server is listening on.
func (s *MetricsServer) Port() int {
	if s == nil {
		return 0
	}
	return s.port
}

// Path returns the path the metrics are served on.
func (s *MetricsServer) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

// URL returns the full URL to access the metrics endpoint.
func (s *MetricsServer) URL() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("http://localhost:%d%s", s.port, s.path)
}
