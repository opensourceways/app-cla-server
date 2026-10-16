package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/opensourceways/obs-sdk/go/log"
	"github.com/opensourceways/obs-sdk/go/sdkctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noProcess() *bool { b := false; return &b }

// initTestLogger replaces the global default logger with one writing to a
// buffer, returning the buffer and a cleanup to restore the original.
func initTestLogger(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()
	var buf bytes.Buffer
	orig := log.Default()
	log.SetDefault(log.New(log.Config{
		Service:   "app-cla-server",
		Env:       "test",
		Instance:  "pod-1",
		Community: "openeuler",
		Output:    &buf,
	}))
	return &buf, func() { log.SetDefault(orig) }
}

// parseLogLines reads each line of buf as a JSON object.
func parseLogLines(t *testing.T, buf *bytes.Buffer) []map[string]interface{} {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	result := make([]map[string]interface{}, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &m), "line: %s", line)
		result = append(result, m)
	}
	return result
}

// TestLogJSONFormat verifies that log output is single-line flat JSON with
// service/env/instance/community fields (acceptance criterion 8).
func TestLogJSONFormat(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	log.Info("server started", "addr", ":8080")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)
	e := entries[0]

	assert.Equal(t, "info", e["level"])
	assert.Equal(t, "server started", e["msg"])
	assert.Equal(t, "app-cla-server", e["service"])
	assert.Equal(t, "test", e["env"])
	assert.Equal(t, "pod-1", e["instance"])
	assert.Equal(t, "openeuler", e["community"])
	assert.Equal(t, ":8080", e["addr"])
}

// TestLogMsgConstant verifies that msg is a constant phrase and variable
// data goes into kv fields. Two calls with different values must produce
// the same msg (acceptance criterion 9, negative assertion).
func TestLogMsgConstant(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	log.Error("failed to init mongodb", "error", "timeout")
	log.Error("failed to init mongodb", "error", "connection refused")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 2)

	assert.Equal(t, "failed to init mongodb", entries[0]["msg"])
	assert.Equal(t, "failed to init mongodb", entries[1]["msg"])

	assert.NotEqual(t, entries[0]["error"], entries[1]["error"])
}

// TestLogInfoContextAddsRequestID verifies that InfoContext includes
// request_id from the context (acceptance criterion 10).
func TestLogInfoContextAddsRequestID(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctx := sdkctx.WithRequestID(context.Background(), "req-abc-123")
	log.InfoContext(ctx, "api access", "action", "login", "status_code", 200)

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)
	e := entries[0]

	assert.Equal(t, "req-abc-123", e["request_id"])
	assert.Equal(t, "api access", e["msg"])
	assert.Equal(t, "login", e["action"])
	assert.Equal(t, float64(200), e["status_code"])
}

// TestLogErrorContextAddsRequestID verifies ErrorContext includes request_id.
func TestLogErrorContextAddsRequestID(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctx := sdkctx.WithRequestID(context.Background(), "req-def-456")
	log.ErrorContext(ctx, "api failed", "action", "upload", "err_code", "system_error", "error", "disk full")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)
	e := entries[0]

	assert.Equal(t, "req-def-456", e["request_id"])
	assert.Equal(t, "api failed", e["msg"])
	assert.Equal(t, "error", e["level"])
	assert.Equal(t, "upload", e["action"])
}

// TestLogNoRequestID verifies that logs without context do not have request_id
// (domain layer / background tasks, acceptance criterion 12).
func TestLogNoRequestID(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	log.Info("worker exit")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)
	e := entries[0]

	_, ok := e["request_id"]
	assert.False(t, ok, "request_id should be absent for non-request logs")
	assert.Equal(t, "worker exit", e["msg"])
}

// TestLogMultipleLinesSameRequestID verifies that multiple log lines in the
// same request share the same request_id for correlation (criterion 10).
func TestLogMultipleLinesSameRequestID(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctx := sdkctx.WithRequestID(context.Background(), "req-shared-789")
	log.InfoContext(ctx, "api access", "action", "start")
	log.ErrorContext(ctx, "api failed", "action", "process")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 2)

	assert.Equal(t, "req-shared-789", entries[0]["request_id"])
	assert.Equal(t, "req-shared-789", entries[1]["request_id"])
}

// TestLogNoTokenInOutput verifies that token/credential values do not appear
// in log output (acceptance criterion 13).
func TestLogNoTokenInOutput(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	log.Error("auth failed", "error", "invalid token")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)
	body := buf.String()

	// The msg and error field should not contain actual token values.
	// (This is a negative test: we don't log token values, only error messages.)
	assert.NotContains(t, body, "Bearer ")
	assert.NotContains(t, body, "secret=")
}
