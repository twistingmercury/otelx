package gin_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twistingmercury/otelx"
	ginmw "github.com/twistingmercury/otelx/middleware/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTelemetry creates a Telemetry instance for testing with a buffer for log capture.
func setupTelemetry(t *testing.T, logBuf *bytes.Buffer) *otelx.Telemetry {
	t.Helper()

	tel, err := otelx.Initialize(context.Background(),
		otelx.WithService("test-service", "1.0.0", "test"),
		otelx.WithLogWriter(logBuf),
		otelx.WithLogLevel(zerolog.DebugLevel),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = tel.Shutdown(context.Background())
	})

	return tel
}

// logEntry represents a parsed log entry for test assertions.
type logEntry struct {
	Level       string  `json:"level"`
	Message     string  `json:"message"`
	Method      string  `json:"method"`
	Path        string  `json:"path"`
	Status      int     `json:"status"`
	LatencyMs   float64 `json:"latency_ms"`
	ClientIP    string  `json:"client_ip"`
	TraceID     string  `json:"trace_id"`
	SpanID      string  `json:"span_id"`
	Query       string  `json:"query"`
	Bytes       int     `json:"bytes"`
	Error       string  `json:"error"`
	XRequestID  string  `json:"x_request_id"`
	Service     string  `json:"service"`
	Version     string  `json:"version"`
	Environment string  `json:"environment"`
}

// parseLogEntry parses the last line of the log buffer as a log entry.
func parseLogEntry(t *testing.T, logBuf *bytes.Buffer) logEntry {
	t.Helper()

	lines := bytes.Split(bytes.TrimSpace(logBuf.Bytes()), []byte("\n"))
	require.NotEmpty(t, lines, "expected at least one log line")

	var entry logEntry
	err := json.Unmarshal(lines[len(lines)-1], &entry)
	require.NoError(t, err, "failed to parse log entry: %s", string(lines[len(lines)-1]))

	return entry
}

func TestLoggingMiddleware_BasicRequest(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	entry := parseLogEntry(t, logBuf)
	assert.Equal(t, "info", entry.Level)
	assert.Equal(t, "request completed", entry.Message)
	assert.Equal(t, "GET", entry.Method)
	assert.Equal(t, "/test", entry.Path)
	assert.Equal(t, http.StatusOK, entry.Status)
	assert.Greater(t, entry.LatencyMs, 0.0)
	assert.NotEmpty(t, entry.ClientIP)
	assert.Equal(t, "test-service", entry.Service)
}

func TestLoggingMiddleware_ErrorStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		expectedLevel string
	}{
		{"client error", http.StatusBadRequest, "warn"},
		{"not found", http.StatusNotFound, "warn"},
		{"server error", http.StatusInternalServerError, "error"},
		{"service unavailable", http.StatusServiceUnavailable, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuf := &bytes.Buffer{}
			tel := setupTelemetry(t, logBuf)

			r := gin.New()
			r.Use(ginmw.LoggingMiddleware(tel))
			r.GET("/test", func(c *gin.Context) {
				c.Status(tt.status)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)

			entry := parseLogEntry(t, logBuf)
			assert.Equal(t, tt.expectedLevel, entry.Level)
			assert.Equal(t, tt.status, entry.Status)
		})
	}
}

func TestLoggingMiddleware_WithSkipPaths(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel,
		ginmw.WithSkipPaths("/health", "/ready"),
	))
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})
	r.GET("/api/users", func(c *gin.Context) {
		c.String(http.StatusOK, "users")
	})

	// Request to skipped path should not log
	logBuf.Reset()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, logBuf.String(), "expected no log for skipped path")

	// Request to non-skipped path should log
	logBuf.Reset()
	req = httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	entry := parseLogEntry(t, logBuf)
	assert.Equal(t, "/api/users", entry.Path)
}

func TestLoggingMiddleware_WithLogLevel(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel,
		ginmw.WithLogLevel(zerolog.DebugLevel),
	))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	entry := parseLogEntry(t, logBuf)
	assert.Equal(t, "debug", entry.Level)
}

func TestLoggingMiddleware_WithRequestHeaders(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel,
		ginmw.WithRequestHeaders("X-Request-ID", "X-Correlation-ID"),
	))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "req-123")
	req.Header.Set("X-Correlation-ID", "corr-456")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	entry := parseLogEntry(t, logBuf)
	assert.Equal(t, "req-123", entry.XRequestID)

	// Check correlation ID manually since it's not in our struct
	var rawEntry map[string]interface{}
	lines := bytes.Split(bytes.TrimSpace(logBuf.Bytes()), []byte("\n"))
	err := json.Unmarshal(lines[len(lines)-1], &rawEntry)
	require.NoError(t, err)
	assert.Equal(t, "corr-456", rawEntry["x_correlation_id"])
}

func TestLoggingMiddleware_WithCustomFields(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel,
		ginmw.WithCustomFields(func(c *gin.Context) map[string]interface{} {
			return map[string]interface{}{
				"user_id":    c.GetString("user_id"),
				"tenant":     "acme",
				"request_no": 42,
			}
		}),
	))
	r.GET("/test", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var rawEntry map[string]interface{}
	lines := bytes.Split(bytes.TrimSpace(logBuf.Bytes()), []byte("\n"))
	err := json.Unmarshal(lines[len(lines)-1], &rawEntry)
	require.NoError(t, err)

	assert.Equal(t, "user-123", rawEntry["user_id"])
	assert.Equal(t, "acme", rawEntry["tenant"])
	assert.Equal(t, float64(42), rawEntry["request_no"])
}

func TestLoggingMiddleware_WithQueryParams(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel))
	r.GET("/search", func(c *gin.Context) {
		c.String(http.StatusOK, "results")
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=test&page=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	entry := parseLogEntry(t, logBuf)
	assert.Equal(t, "q=test&page=1", entry.Query)
}

func TestLoggingMiddleware_WithGinErrors(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel))
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(gin.Error{
			Err:  assert.AnError,
			Type: gin.ErrorTypePublic,
			Meta: "test error",
		})
		c.Status(http.StatusBadRequest)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	entry := parseLogEntry(t, logBuf)
	assert.Contains(t, entry.Error, "assert.AnError")
}

func TestCorrelationMiddleware(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.CorrelationMiddleware(tel))

	var capturedLogger zerolog.Logger
	r.GET("/test", func(c *gin.Context) {
		capturedLogger = ginmw.Logger(c, tel)
		capturedLogger.Info().Msg("handler log")
		c.String(http.StatusOK, "ok")
	})

	// Reset buffer to check only handler logs
	logBuf.Reset()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should only have the handler log, not the request completion log
	lines := bytes.Split(bytes.TrimSpace(logBuf.Bytes()), []byte("\n"))
	assert.Len(t, lines, 1, "expected only handler log, no request completion log")

	entry := parseLogEntry(t, logBuf)
	assert.Equal(t, "handler log", entry.Message)
}

func TestLogger_WithMiddleware(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel))

	var capturedLogger zerolog.Logger
	r.GET("/test", func(c *gin.Context) {
		capturedLogger = ginmw.Logger(c, tel)
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Logger should be usable and have service context
	capturedLogger.Info().Msg("test log after request")

	var rawEntry map[string]interface{}
	lines := bytes.Split(bytes.TrimSpace(logBuf.Bytes()), []byte("\n"))
	err := json.Unmarshal(lines[len(lines)-1], &rawEntry)
	require.NoError(t, err)

	assert.Equal(t, "test log after request", rawEntry["message"])
	assert.Equal(t, "test-service", rawEntry["service"])
}

func TestLogger_WithoutMiddleware(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	// Note: No middleware applied

	var capturedLogger zerolog.Logger
	r.GET("/test", func(c *gin.Context) {
		// Logger should fall back to tel.Logger
		capturedLogger = ginmw.Logger(c, tel)
		capturedLogger.Info().Msg("fallback log")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	entry := parseLogEntry(t, logBuf)
	assert.Equal(t, "fallback log", entry.Message)
	assert.Equal(t, "test-service", entry.Service)
}

func TestLogger_WithNilTelemetry(t *testing.T) {
	r := gin.New()

	r.GET("/test", func(c *gin.Context) {
		// Should return a no-op logger when tel is nil
		logger := ginmw.Logger(c, nil)
		// Log something to verify it doesn't panic (no-op logger)
		logger.Info().Msg("this should not panic")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should not panic
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMustLogger_WithMiddleware(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel))

	var capturedLogger zerolog.Logger
	r.GET("/test", func(c *gin.Context) {
		capturedLogger = ginmw.MustLogger(c)
		capturedLogger.Info().Msg("must logger test")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMustLogger_WithoutMiddleware_Panics(t *testing.T) {
	r := gin.New()
	// Note: No middleware applied

	r.GET("/test", func(c *gin.Context) {
		// Should panic because middleware is not applied
		_ = ginmw.MustLogger(c)
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	assert.Panics(t, func() {
		r.ServeHTTP(w, req)
	})
}

func TestLoggingMiddleware_AllOptions(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel,
		ginmw.WithSkipPaths("/health"),
		ginmw.WithLogLevel(zerolog.DebugLevel),
		ginmw.WithRequestHeaders("X-Request-ID"),
		ginmw.WithCustomFields(func(c *gin.Context) map[string]interface{} {
			return map[string]interface{}{"custom": "field"}
		}),
	))
	r.GET("/api/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("X-Request-ID", "req-all-opts")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var rawEntry map[string]interface{}
	lines := bytes.Split(bytes.TrimSpace(logBuf.Bytes()), []byte("\n"))
	err := json.Unmarshal(lines[len(lines)-1], &rawEntry)
	require.NoError(t, err)

	assert.Equal(t, "debug", rawEntry["level"])
	assert.Equal(t, "req-all-opts", rawEntry["x_request_id"])
	assert.Equal(t, "field", rawEntry["custom"])
}

func TestLoggingMiddleware_ResponseSize(t *testing.T) {
	logBuf := &bytes.Buffer{}
	tel := setupTelemetry(t, logBuf)

	r := gin.New()
	r.Use(ginmw.LoggingMiddleware(tel))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	entry := parseLogEntry(t, logBuf)
	assert.Equal(t, len("hello world"), entry.Bytes)
}

func TestHeaderToFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"X-Request-ID", "x_request_id"},
		{"X-Correlation-ID", "x_correlation_id"},
		{"Content-Type", "content_type"},
		{"Accept", "accept"},
		{"X-CUSTOM-HEADER", "x_custom_header"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			logBuf := &bytes.Buffer{}
			tel := setupTelemetry(t, logBuf)

			r := gin.New()
			r.Use(ginmw.LoggingMiddleware(tel,
				ginmw.WithRequestHeaders(tt.input),
			))
			r.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set(tt.input, "test-value")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			var rawEntry map[string]interface{}
			lines := bytes.Split(bytes.TrimSpace(logBuf.Bytes()), []byte("\n"))
			err := json.Unmarshal(lines[len(lines)-1], &rawEntry)
			require.NoError(t, err)

			assert.Equal(t, "test-value", rawEntry[tt.expected])
		})
	}
}
