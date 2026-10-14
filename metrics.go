package otelx

import (
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

// metricsServer serves Prometheus metrics for one Telemetry.
type metricsServer struct {
	server *http.Server
	url    string
}

// setupMetrics returns a MeterProvider and a running metrics server, or nils
// when metrics are disabled. It binds cfg.MetricsAddr before returning, so a
// port conflict is reported as an error.
//
// Each call uses its own Prometheus registry rather than the global default,
// so several Telemetry instances in one process do not collide.
func setupMetrics(cfg Config, res *resource.Resource) (*metric.MeterProvider, *metricsServer, error) {
	if cfg.MetricsAddr == "" {
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

	listener, err := net.Listen("tcp", cfg.MetricsAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on metrics address %q: %w", cfg.MetricsAddr, err)
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

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("metrics server error: %v\n", err)
		}
	}()

	return mp, &metricsServer{
		server: server,
		url:    metricsURL(listener.Addr(), cfg.MetricsPath),
	}, nil
}

// metricsURL builds a URL for addr, using localhost when the server listens
// on every interface.
func metricsURL(addr net.Addr, path string) string {
	host, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		return ""
	}
	if ip := net.ParseIP(host); ip == nil || ip.IsUnspecified() {
		host = "localhost"
	}
	return "http://" + net.JoinHostPort(host, port) + path
}
