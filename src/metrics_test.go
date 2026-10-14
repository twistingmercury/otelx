package otelx_test

import (
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twistingmercury/otelx/src/v2"
	"github.com/twistingmercury/otelx/src/v2/internal/testutil"
)

func initMetrics(t *testing.T, cfg otelx.Config) *otelx.Telemetry {
	t.Helper()
	cfg.ServiceName = "svc"
	cfg.LogWriter = testutil.NopWriter{}
	if cfg.MetricsAddr == "" {
		cfg.MetricsAddr = "127.0.0.1:0"
	}
	tel, err := otelx.Initialize(context.Background(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tel.Shutdown(context.Background()) })
	return tel
}

func TestMetrics_ServesCounters(t *testing.T) {
	ctx := context.Background()
	tel := initMetrics(t, otelx.Config{})

	counter, err := tel.MeterProvider.Meter("test").Int64Counter("requests")
	require.NoError(t, err)
	counter.Add(ctx, 3)

	body, status := testutil.FetchURL(t, tel.MetricsURL())
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, "requests_total")
	assert.Contains(t, body, `service_name="svc"`)
}

func TestMetrics_CustomPath(t *testing.T) {
	tel := initMetrics(t, otelx.Config{MetricsPath: "/custom"})
	require.Regexp(t, `^http://127\.0\.0\.1:\d+/custom$`, tel.MetricsURL())

	_, status := testutil.FetchURL(t, tel.MetricsURL())
	assert.Equal(t, http.StatusOK, status)
}

func TestMetrics_URLUsesLocalhostForWildcardAddress(t *testing.T) {
	tel := initMetrics(t, otelx.Config{MetricsAddr: ":0"})
	assert.Regexp(t, `^http://localhost:\d+/metrics$`, tel.MetricsURL())
}

func TestInitialize_MetricsPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()

	tel, err := otelx.Initialize(context.Background(), otelx.Config{
		ServiceName: "svc",
		LogWriter:   testutil.NopWriter{},
		MetricsAddr: ln.Addr().String(),
	})
	require.Error(t, err, "Initialize must fail when the metrics port is taken")
	assert.Nil(t, tel)
}

// Two Telemetry instances in one process must each serve their own metrics.
func TestInitialize_TwiceServesSeparateMetrics(t *testing.T) {
	ctx := context.Background()

	init := func(counter string) *otelx.Telemetry {
		tel := initMetrics(t, otelx.Config{})
		c, err := tel.MeterProvider.Meter("test").Int64Counter(counter)
		require.NoError(t, err)
		c.Add(ctx, 1)
		return tel
	}
	first := init("first_counter")
	second := init("second_counter")

	firstBody, _ := testutil.FetchURL(t, first.MetricsURL())
	secondBody, _ := testutil.FetchURL(t, second.MetricsURL())

	assert.Contains(t, firstBody, "first_counter")
	assert.NotContains(t, firstBody, "second_counter")
	assert.Contains(t, secondBody, "second_counter")
	assert.NotContains(t, secondBody, "first_counter")
	assert.Contains(t, firstBody, "go_goroutines", "Go runtime metrics should still be served")
}

func TestMetrics_StopsOnShutdown(t *testing.T) {
	tel, err := otelx.Initialize(context.Background(), otelx.Config{
		ServiceName: "svc",
		LogWriter:   testutil.NopWriter{},
		MetricsAddr: "127.0.0.1:0",
	})
	require.NoError(t, err)
	url := tel.MetricsURL()
	require.NoError(t, tel.Shutdown(context.Background()))

	_, err = http.Get(url) //nolint:gosec // test URL
	assert.Error(t, err)
}
