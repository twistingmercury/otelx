package otelx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
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
// It binds the metrics port before returning, so a port conflict is reported
// as an error, and then serves the /metrics endpoint in the background.
//
// Each call uses its own Prometheus registry rather than the global default,
// so several Telemetry instances in one process do not collide.
func setupMetrics(_ context.Context, cfg *Config, res *resource.Resource) (*metric.MeterProvider, *MetricsServer, error) {
	if !cfg.MetricsEnabled {
		return nil, nil, nil
	}

	registry := prom.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	exporter, err := prometheus.New(prometheus.WithRegisterer(registry))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.MetricsPort))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on metrics port %d: %w", cfg.MetricsPort, err)
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(exporter),
	)

	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	metricsServer := &MetricsServer{
		server: server,
		port:   cfg.MetricsPort,
		path:   cfg.MetricsPath,
	}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
