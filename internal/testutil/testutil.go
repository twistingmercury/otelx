// Package testutil provides test helpers for the otelx package.
package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

// LogCapture captures log output for testing.
type LogCapture struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

// NewLogCapture creates a new LogCapture instance.
func NewLogCapture() *LogCapture {
	return &LogCapture{}
}

// Write implements io.Writer.
func (lc *LogCapture) Write(p []byte) (n int, err error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.buffer.Write(p)
}

// String returns the captured log output.
func (lc *LogCapture) String() string {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.buffer.String()
}

// Bytes returns the captured log output as bytes.
func (lc *LogCapture) Bytes() []byte {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.buffer.Bytes()
}

// Reset clears the captured log output.
func (lc *LogCapture) Reset() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.buffer.Reset()
}

// Contains checks if the captured log output contains the given substring.
func (lc *LogCapture) Contains(s string) bool {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return bytes.Contains(lc.buffer.Bytes(), []byte(s))
}

// ParseJSONLogs parses the captured log output as JSON lines.
func (lc *LogCapture) ParseJSONLogs() ([]map[string]interface{}, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	var logs []map[string]interface{}
	lines := bytes.Split(lc.buffer.Bytes(), []byte("\n"))

	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var entry map[string]interface{}
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, err
		}
		logs = append(logs, entry)
	}

	return logs, nil
}

// GetFreePort returns an available port for testing.
func GetFreePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	defer func() { _ = listener.Close() }()

	return listener.Addr().(*net.TCPAddr).Port
}

// WaitForServer waits for an HTTP server to become available.
func WaitForServer(t *testing.T, url string, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 100 * time.Millisecond}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}

	return false
}

// FetchURL fetches a URL and returns the response body.
func FetchURL(t *testing.T, url string) (string, int) {
	t.Helper()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("failed to fetch URL %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	return string(body), resp.StatusCode
}

// ContextWithTimeout creates a context with a timeout for testing.
func ContextWithTimeout(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

// RequireEventually retries a condition until it succeeds or times out.
func RequireEventually(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration, msg string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}

	t.Fatalf("condition not met within %v: %s", timeout, msg)
}

// NopWriter is a writer that discards all input.
type NopWriter struct{}

// Write implements io.Writer.
func (NopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
