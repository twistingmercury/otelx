package otelx_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twistingmercury/otelx"
	"github.com/twistingmercury/otelx/internal/testutil"
)

func TestMetricsServer_URL(t *testing.T) {
	// Test nil server
	var nilServer *otelx.MetricsServer
	assert.Equal(t, "", nilServer.URL())
	assert.Equal(t, 0, nilServer.Port())
	assert.Equal(t, "", nilServer.Path())
}

func TestMetrics_Integration(t *testing.T) {
	ctx := context.Background()
	port := testutil.GetFreePort(t)

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithMetrics(port),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	// Wait for server to start
	metricsURL := tel.MetricsURL()
	require.NotEmpty(t, metricsURL)

	ok := testutil.WaitForServer(t, metricsURL, 5*time.Second)
	require.True(t, ok, "metrics server did not start in time")

	// Fetch metrics
	resp, err := http.Get(metricsURL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMetrics_CustomPath(t *testing.T) {
	ctx := context.Background()
	port := testutil.GetFreePort(t)

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithMetrics(port),
		otelx.WithMetricsPath("/custom/metrics"),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	// Wait for server to start
	metricsURL := tel.MetricsURL()
	require.Contains(t, metricsURL, "/custom/metrics")

	ok := testutil.WaitForServer(t, metricsURL, 5*time.Second)
	require.True(t, ok, "metrics server did not start in time")

	// Fetch metrics
	resp, err := http.Get(metricsURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMetrics_Disabled(t *testing.T) {
	ctx := context.Background()

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithoutLogging(),
		// No WithMetrics - metrics disabled
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	assert.Nil(t, tel.MeterProvider)
	assert.Empty(t, tel.MetricsURL())
}

func TestMetrics_Counter(t *testing.T) {
	ctx := context.Background()
	port := testutil.GetFreePort(t)

	tel, err := otelx.Initialize(ctx,
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithMetrics(port),
		otelx.WithoutLogging(),
	)
	require.NoError(t, err)
	require.NotNil(t, tel)
	defer tel.Shutdown(ctx)

	// Wait for server to start
	ok := testutil.WaitForServer(t, tel.MetricsURL(), 5*time.Second)
	require.True(t, ok, "metrics server did not start in time")

	// Create and use a counter
	require.NotNil(t, tel.MeterProvider)
	meter := tel.MeterProvider.Meter("test")
	counter, err := meter.Int64Counter("test_requests_total")
	require.NoError(t, err)

	counter.Add(ctx, 5)

	// Give time for metrics to be collected
	time.Sleep(100 * time.Millisecond)

	// Fetch and verify metrics contain our counter
	body, status := testutil.FetchURL(t, tel.MetricsURL())
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, "test_requests_total")
}
