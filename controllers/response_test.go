package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	beecontext "github.com/beego/beego/v2/server/web/context"
	"github.com/opensourceways/obs-sdk/go/log"
	"github.com/opensourceways/obs-sdk/go/sdkctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// newBaseControllerWithRequest builds a baseController backed by a real
// http.Request, so that Ctx.Request.Context() carries sdkctx values
// injected by the middleware.
func newBaseControllerWithRequest(reqID string) *baseController {
	ctx := beecontext.NewContext()

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/v1/link/abc", nil)
	inner := sdkctx.WithRequestID(context.Background(), reqID)
	r = r.WithContext(inner)

	ctx.Request = r
	ctx.ResponseWriter = &beecontext.Response{
		ResponseWriter: recorder,
	}
	ctx.Input.Reset(ctx)
	ctx.Output.Context = ctx

	ctl := &baseController{}
	ctl.Ctx = ctx
	ctl.Data = map[interface{}]interface{}{}

	return ctl
}

// TestOperationLogAlwaysLogs verifies that operationLog produces a log line
// for every API response, including failure paths (400/401/403) where user
// is empty (acceptance criterion 11).
func TestOperationLogAlwaysLogs(t *testing.T) {
	tests := []struct {
		name       string
		action     string
		statusCode int
	}{
		{"400 missing param", "create corp signing", 400},
		{"401 unauthorized", "access resource", 401},
		{"403 forbidden", "upload pdf", 403},
		{"200 success", "get cla", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, cleanup := initTestLogger(t)
			defer cleanup()

			ctl := newBaseControllerWithRequest("req-" + tt.name)
			ctl.operationLog(tt.action, tt.statusCode)

			entries := parseLogLines(t, buf)
			require.Len(t, entries, 1)
			e := entries[0]

			assert.Equal(t, "api access", e["msg"])
			assert.Equal(t, tt.action, e["action"])
			assert.Equal(t, float64(tt.statusCode), e["status_code"])
			assert.Equal(t, "req-"+tt.name, e["request_id"])
			assert.Equal(t, "", e["user"])
		})
	}
}

// TestAddOperationLog verifies addOperationLog produces a log line with
// user, action, and status_code (acceptance criterion 11).
func TestAddOperationLog(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-add-op")
	ctl.addOperationLog("manager1", "reset password", 200)

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)
	e := entries[0]

	assert.Equal(t, "api access", e["msg"])
	assert.Equal(t, "manager1", e["user"])
	assert.Equal(t, "reset password", e["action"])
	assert.Equal(t, float64(200), e["status_code"])
	assert.Equal(t, "req-add-op", e["request_id"])
}

// TestSendFailedResponseLogs500 verifies that 5xx failures produce an error
// log with action, err_code, and error, plus an access log with status_code
// (acceptance criterion 11).
func TestSendFailedResponseLogs500(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-500")
	ctl.sendFailedResponse(500, "system_error", errLogTest{}, "process signing")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 2)

	// First entry: error log from sendFailedResponse.
	e1 := entries[0]
	assert.Equal(t, "error", e1["level"])
	assert.Equal(t, "api failed", e1["msg"])
	assert.Equal(t, "process signing", e1["action"])
	assert.Equal(t, "req-500", e1["request_id"])

	// Second entry: access log from operationLog via sendResponse.
	e2 := entries[1]
	assert.Equal(t, "info", e2["level"])
	assert.Equal(t, "api access", e2["msg"])
	assert.Equal(t, "process signing", e2["action"])
	assert.Equal(t, float64(500), e2["status_code"])
}

type errLogTest struct{}

func (e errLogTest) Error() string { return "internal failure" }

// TestOperationLogJSONFormat verifies the output is single-line flat JSON
// with service/env/instance/community (acceptance criterion 8).
func TestOperationLogJSONFormat(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-fmt")
	ctl.operationLog("get link", 200)

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)
	e := entries[0]

	assert.Equal(t, "app-cla-server", e["service"])
	assert.Equal(t, "test", e["env"])
	assert.Equal(t, "pod-1", e["instance"])
	assert.Equal(t, "openeuler", e["community"])
	assert.Equal(t, "info", e["level"])
}

// TestOperationLogWithUser verifies that operationLog extracts the user
// identity from the access controller when present (covering the 401 path
// where user is absent and the 200 path where user is present).
func TestOperationLogWithUser(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-user")
	ctl.Data[apiAccessController] = accessController{
		RemoteAddr: "1.2.3.4",
		Permission: PermissionOwnerOfOrg,
		Payload:    &acForCorpManagerPayload{UserId: "user1"},
	}
	ctl.operationLog("get signing", 200)

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)
	e := entries[0]

	assert.Equal(t, "user1", e["user"])
	assert.Equal(t, "get signing", e["action"])
	assert.Equal(t, float64(200), e["status_code"])
}
