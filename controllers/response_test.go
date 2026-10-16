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

// TestOperationLogCondition verifies that operationLog only writes a log line
// when user is non-empty. When user is empty (401/403/400 with no access
// controller), no log line is produced.
func TestOperationLogCondition(t *testing.T) {
	t.Run("empty user produces no log", func(t *testing.T) {
		buf, cleanup := initTestLogger(t)
		defer cleanup()

		ctl := newBaseControllerWithRequest("req-empty")
		ctl.operationLog("access resource", 401)

		entries := parseLogLines(t, buf)
		require.Len(t, entries, 0, "no log line should be produced when user is empty")
	})

	t.Run("non-empty user produces valid JSON log", func(t *testing.T) {
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

		assert.Equal(t, "api access", e["msg"])
		assert.Equal(t, "user1", e["user"])
		assert.Equal(t, "get signing", e["action"])
		assert.Equal(t, float64(200), e["status_code"])
		assert.Equal(t, "req-user", e["request_id"])
	})
}

// TestAddOperationLog verifies addOperationLog produces a log line with
// user, action, and status_code.
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
// log with action, err_code, and error. When user is present, an access log
// with status_code is also produced via operationLog.
func TestSendFailedResponseLogs500(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-500")
	ctl.Data[apiAccessController] = accessController{
		RemoteAddr: "1.2.3.4",
		Permission: PermissionOwnerOfOrg,
		Payload:    &acForCorpManagerPayload{UserId: "admin1"},
	}
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
	assert.Equal(t, "admin1", e2["user"])
}

// TestSendFailedResponse500NoUserLogsOnlyError verifies that when user is
// empty, a 5xx response only produces the error log (no access log).
func TestSendFailedResponse500NoUserLogsOnlyError(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-500-nouser")
	ctl.sendFailedResponse(500, "system_error", errLogTest{}, "process signing")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "error", e["level"])
	assert.Equal(t, "api failed", e["msg"])
	assert.Equal(t, "process signing", e["action"])
	assert.Equal(t, "req-500-nouser", e["request_id"])
}

type errLogTest struct{}

func (e errLogTest) Error() string { return "internal failure" }

// TestSendFailedResponseLogs400 verifies that 400 failures produce a warn
// log with action, status_code, err_code, and error.
func TestSendFailedResponseLogs400(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-400")
	ctl.sendFailedResponse(400, "missing_url_path_parameter", errLogTest{}, "access resource")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "warn", e["level"])
	assert.Equal(t, "api failed", e["msg"])
	assert.Equal(t, "access resource", e["action"])
	assert.Equal(t, float64(400), e["status_code"])
	assert.Equal(t, "missing_url_path_parameter", e["err_code"])
	assert.Equal(t, "internal failure", e["error"])
	assert.Equal(t, "req-400", e["request_id"])
}

// TestSendFailedResponseLogs401 verifies that 401 failures produce a warn
// log with action, status_code, err_code, and error.
func TestSendFailedResponseLogs401(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-401")
	ctl.sendFailedResponse(401, "unknown_token", errLogTest{}, "access resource")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "warn", e["level"])
	assert.Equal(t, "api failed", e["msg"])
	assert.Equal(t, "access resource", e["action"])
	assert.Equal(t, float64(401), e["status_code"])
	assert.Equal(t, "unknown_token", e["err_code"])
	assert.Equal(t, "internal failure", e["error"])
	assert.Equal(t, "req-401", e["request_id"])
}

// TestSendFailedResponseLogs403 verifies that 403 failures produce a warn
// log with action, status_code, err_code, and error.
func TestSendFailedResponseLogs403(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-403")
	ctl.sendFailedResponse(403, "unauthorized_token", errLogTest{}, "access resource")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "warn", e["level"])
	assert.Equal(t, "api failed", e["msg"])
	assert.Equal(t, "access resource", e["action"])
	assert.Equal(t, float64(403), e["status_code"])
	assert.Equal(t, "unauthorized_token", e["err_code"])
	assert.Equal(t, "internal failure", e["error"])
	assert.Equal(t, "req-403", e["request_id"])
}

// TestSendFailedResponse4xxWithUserLogsWarnAndAccess verifies that when user
// is present, a 4xx failure produces both a warn log and an access log.
func TestSendFailedResponse4xxWithUserLogsWarnAndAccess(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-400-user")
	ctl.Data[apiAccessController] = accessController{
		RemoteAddr: "1.2.3.4",
		Permission: PermissionOwnerOfOrg,
		Payload:    &acForCorpManagerPayload{UserId: "admin1"},
	}
	ctl.sendFailedResponse(403, "unauthorized_token", errLogTest{}, "process signing")

	entries := parseLogLines(t, buf)
	require.Len(t, entries, 2)

	e1 := entries[0]
	assert.Equal(t, "warn", e1["level"])
	assert.Equal(t, "api failed", e1["msg"])
	assert.Equal(t, "process signing", e1["action"])
	assert.Equal(t, float64(403), e1["status_code"])

	e2 := entries[1]
	assert.Equal(t, "info", e2["level"])
	assert.Equal(t, "api access", e2["msg"])
	assert.Equal(t, "process signing", e2["action"])
	assert.Equal(t, float64(403), e2["status_code"])
	assert.Equal(t, "admin1", e2["user"])
}

// TestOperationLogJSONFormat verifies the output is single-line flat JSON
// with service/env/instance/community (acceptance criterion 8).
func TestOperationLogJSONFormat(t *testing.T) {
	buf, cleanup := initTestLogger(t)
	defer cleanup()

	ctl := newBaseControllerWithRequest("req-fmt")
	ctl.Data[apiAccessController] = accessController{
		RemoteAddr: "1.2.3.4",
		Permission: PermissionOwnerOfOrg,
		Payload:    &acForCorpManagerPayload{UserId: "user1"},
	}
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
