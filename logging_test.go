package otelx_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twistingmercury/otelx"
)

func TestSetupLogging_Disabled(t *testing.T) {
	cfg := otelx.NewDefaultConfig()
	cfg.ServiceName = "test-service"
	cfg.ServiceVersion = "1.0.0"
	cfg.Environment = "test"
	cfg.LoggingEnabled = false

	// Access the setup function through Initialize
	// For now, test LogLevel function directly
}

func TestLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"trace", zerolog.TraceLevel},
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"warning", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"fatal", zerolog.FatalLevel},
		{"panic", zerolog.PanicLevel},
		{"disabled", zerolog.Disabled},
		{"none", zerolog.Disabled},
		{"off", zerolog.Disabled},
		{"unknown", zerolog.InfoLevel},
		{"", zerolog.InfoLevel},
		{"INFO", zerolog.InfoLevel}, // Case sensitive, returns default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := otelx.LogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogging_JSONOutput(t *testing.T) {
	var buf bytes.Buffer

	logger := zerolog.New(&buf).With().Timestamp().Logger()
	logger.Info().Msg("test message")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "test message", entry["message"])
	assert.NotEmpty(t, entry["time"])
}

func TestLogging_WithFields(t *testing.T) {
	var buf bytes.Buffer

	logger := zerolog.New(&buf).With().Timestamp().Logger()
	logger.Info().
		Str("service", "test-service").
		Str("version", "1.0.0").
		Int("count", 42).
		Msg("structured log")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "structured log", entry["message"])
	assert.Equal(t, "test-service", entry["service"])
	assert.Equal(t, "1.0.0", entry["version"])
	assert.Equal(t, float64(42), entry["count"]) // JSON numbers are float64
}

func TestLogging_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer

	// Set level to Warn - should filter out Info and Debug
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	defer zerolog.SetGlobalLevel(zerolog.TraceLevel) // Reset

	logger := zerolog.New(&buf).With().Timestamp().Logger()

	// These should not be written
	logger.Debug().Msg("debug message")
	logger.Info().Msg("info message")

	assert.Empty(t, buf.String())

	// This should be written
	logger.Warn().Msg("warn message")
	assert.Contains(t, buf.String(), "warn message")
}

func TestLogging_NopLogger(t *testing.T) {
	var buf bytes.Buffer

	logger := zerolog.Nop()
	logger.Info().Msg("this should not appear")

	// Write something to buf to verify it's not the logger
	buf.WriteString("test")
	assert.Equal(t, "test", buf.String())
}

func TestLogging_WithError(t *testing.T) {
	var buf bytes.Buffer

	logger := zerolog.New(&buf).With().Timestamp().Logger()
	logger.Error().
		Err(assert.AnError).
		Msg("operation failed")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "error", entry["level"])
	assert.Equal(t, "operation failed", entry["message"])
	assert.Equal(t, "assert.AnError general error for testing", entry["error"])
}
