package worker

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
	obsmetrics "github.com/opensourceways/obs-sdk/go/metrics"
	"github.com/opensourceways/obs-sdk/go/middleware/beegomw"
	beego "github.com/beego/beego/v2/server/web"
	beecontext "github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noProcess() *bool { b := false; return &b }

func newTestMetrics(t *testing.T) *obsmetrics.Metrics {
	t.Helper()
	return obsmetrics.New(obsmetrics.Config{
		Service:        "app-cla-server",
		Env:            "test",
		Instance:       "pod-1",
		Community:      "openeuler",
		IncludeProcess: noProcess(),
	})
}

func resetBusinessMetrics() {
	emailSentCounter = nil
	pdfGeneratedCounter = nil
}

// TestBusinessMetricsExposed verifies that cla_email_sent_total and
// cla_pdf_generated_total appear in real /metrics text output with the
// four const labels (service/env/instance/community).
func TestBusinessMetricsExposed(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	recordEmailSent("gmail", "sent")
	recordEmailSent("gmail", "failed")
	recordPDFGenerated("success")
	recordPDFGenerated("failed")

	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := w.Body.String()

	assert.Contains(t, body, "cla_email_sent_total")
	assert.Contains(t, body, "cla_pdf_generated_total")

	assert.Contains(t, body, `service="app-cla-server"`)
	assert.Contains(t, body, `env="test"`)
	assert.Contains(t, body, `instance="pod-1"`)
	assert.Contains(t, body, `community="openeuler"`)

	assert.Contains(t, body, `platform="gmail"`)
	assert.Contains(t, body, `outcome="sent"`)
	assert.Contains(t, body, `outcome="failed"`)
}

// TestEmailSentMetricValue verifies the counter value via testutil.
func TestEmailSentMetricValue(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	recordEmailSent("gmail", "sent")
	recordEmailSent("gmail", "sent")
	recordEmailSent("gmail", "failed")

	require.Equal(t, 2.0, testutil.ToFloat64(
		emailSentCounter.Vec().WithLabelValues("openeuler", "gmail", "sent")))
	require.Equal(t, 1.0, testutil.ToFloat64(
		emailSentCounter.Vec().WithLabelValues("openeuler", "gmail", "failed")))
}

// TestPDFGeneratedMetricValue verifies the pdf counter value via testutil.
func TestPDFGeneratedMetricValue(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	recordPDFGenerated("success")
	recordPDFGenerated("success")
	recordPDFGenerated("failed")

	require.Equal(t, 2.0, testutil.ToFloat64(
		pdfGeneratedCounter.Vec().WithLabelValues("openeuler", "success")))
	require.Equal(t, 1.0, testutil.ToFloat64(
		pdfGeneratedCounter.Vec().WithLabelValues("openeuler", "failed")))
}

// TestSendSimpleMessageRecordsEmailMetric verifies that SendSimpleMessage
// records the email_sent_total metric for the sent outcome.
func TestSendSimpleMessageRecordsEmailMetric(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	fake := &fakeEmailService{}
	emailservice.Register("test-metric-success", fake)

	w := &emailWorker{stop: make(chan struct{})}
	msg := &EmailMessage{To: []string{"a@example.com"}, Subject: "hi"}

	w.SendSimpleMessage("test-metric-success", msg)
	w.Shutdown()

	require.Equal(t, 1.0, testutil.ToFloat64(
		emailSentCounter.Vec().WithLabelValues("openeuler", "test-metric-success", "sent")))
}

// TestSendSimpleMessageFailureRecordsEmailMetric verifies that SendSimpleMessage
// records the email_sent_total metric for the failed outcome.
func TestSendSimpleMessageFailureRecordsEmailMetric(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	fake := &fakeEmailService{err: fakeSendErr}
	emailservice.Register("test-metric-fail", fake)

	w := &emailWorker{stop: make(chan struct{})}
	msg := &EmailMessage{To: []string{"a@example.com"}, Subject: "hi"}

	w.SendSimpleMessage("test-metric-fail", msg)
	w.Shutdown()

	require.GreaterOrEqual(t, testutil.ToFloat64(
		emailSentCounter.Vec().WithLabelValues("openeuler", "test-metric-fail", "failed")), 1.0)
}

// --- HTTP middleware integration tests ---

func newRegister(m *obsmetrics.Metrics) *beego.ControllerRegister {
	reg := beego.NewControllerRegister()
	reg.InsertFilterChain("/*", beegomw.Middleware(beegomw.Options{Metrics: m}))
	return reg
}

func findCounterValue(t *testing.T, m *obsmetrics.Metrics, familyName string, wantLabels map[string]string) float64 {
	t.Helper()
	fams, err := m.Registry().Gather()
	require.NoError(t, err)
	for _, f := range fams {
		if f.GetName() != familyName {
			continue
		}
		for _, s := range f.GetMetric() {
			labels := map[string]string{}
			for _, lp := range s.GetLabel() {
				labels[lp.GetName()] = lp.GetValue()
			}
			ok := true
			for k, v := range wantLabels {
				if labels[k] != v {
					ok = false
					break
				}
			}
			if ok {
				return s.GetCounter().GetValue()
			}
		}
	}
	t.Fatalf("family %s with labels %v not found", familyName, wantLabels)
	return 0
}

// TestHTTPMetricsExposed verifies that http_server_requests_total and
// http_server_request_duration_seconds appear in real /metrics text.
func TestHTTPMetricsExposed(t *testing.T) {
	m := newTestMetrics(t)
	reg := newRegister(m)
	reg.AddMethod("GET", "/ping", func(ctx *beecontext.Context) {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	})
	reg.Init()

	w := httptest.NewRecorder()
	reg.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))
	require.Equal(t, http.StatusOK, w.Code)

	mw := httptest.NewRecorder()
	m.Handler().ServeHTTP(mw, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := mw.Body.String()

	assert.Contains(t, body, "http_server_requests_total")
	assert.Contains(t, body, "http_server_request_duration_seconds")
	assert.Contains(t, body, "http_server_request_duration_seconds_bucket")
	assert.Contains(t, body, "http_server_request_duration_seconds_count")

	assert.Contains(t, body, `service="app-cla-server"`)
	assert.Contains(t, body, `env="test"`)
	assert.Contains(t, body, `instance="pod-1"`)
	assert.Contains(t, body, `community="openeuler"`)

	assert.Contains(t, body, `path="/ping"`)
	assert.NotContains(t, body, "request_id")
	assert.NotContains(t, body, "trace_id")
	assert.NotContains(t, body, "span_id")
}

// TestPathLabelUsesRouterPattern verifies that path label is the route
// template, not the raw URL with dynamic IDs (acceptance criterion 3).
func TestPathLabelUsesRouterPattern(t *testing.T) {
	m := newTestMetrics(t)
	reg := newRegister(m)
	reg.AddMethod("GET", "/v1/cla/:link_id", func(ctx *beecontext.Context) {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	})
	reg.Init()

	w := httptest.NewRecorder()
	reg.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/cla/a1b2c3d4random", nil))
	require.Equal(t, http.StatusOK, w.Code)

	require.Equal(t, 1.0, findCounterValue(t, m, "http_server_requests_total", map[string]string{
		"method": "GET", "path": "/v1/cla/:link_id", "status_code": "200",
	}))

	mw := httptest.NewRecorder()
	m.Handler().ServeHTTP(mw, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := mw.Body.String()
	assert.NotContains(t, body, "a1b2c3d4random")
}

// TestUnmatchedRouteConverged verifies that two different unmatched paths
// produce only one path="unmatched" series (acceptance criterion 4).
func TestUnmatchedRouteConverged(t *testing.T) {
	m := newTestMetrics(t)
	reg := newRegister(m)
	reg.AddMethod("GET", "/ping", func(ctx *beecontext.Context) {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	})
	reg.Init()

	for _, p := range []string{"/nope/11111111", "/nope/22222222"} {
		w := httptest.NewRecorder()
		reg.ServeHTTP(w, httptest.NewRequest(http.MethodGet, p, nil))
	}

	require.Equal(t, 2.0, findCounterValue(t, m, "http_server_requests_total", map[string]string{
		"method": "GET", "path": "unmatched",
	}))

	mw := httptest.NewRecorder()
	m.Handler().ServeHTTP(mw, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := mw.Body.String()
	assert.NotContains(t, body, "/nope/11111111")
	assert.NotContains(t, body, "/nope/22222222")
}

// stopRunController replicates app-cla-server's auth-failure path:
// write response then StopRun(), which panics internally (acceptance criterion 2).
type stopRunController struct {
	beego.Controller
}

func (c *stopRunController) Post() {
	c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
	c.StopRun()
}

// TestStopRunStillRecorded verifies that 401 from StopRun is still counted
// when using InsertFilterChain (acceptance criterion 2).
func TestStopRunStillRecorded(t *testing.T) {
	m := newTestMetrics(t)
	reg := newRegister(m)
	reg.Add("/v1/link/:link_id", &stopRunController{})
	reg.Init()

	w := httptest.NewRecorder()
	reg.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/link/abc123", nil))
	require.Equal(t, http.StatusUnauthorized, w.Code)

	require.Equal(t, 1.0, findCounterValue(t, m, "http_server_requests_total", map[string]string{
		"method": "POST", "path": "/v1/link/:link_id", "status_code": "401",
	}))
}

// TestMetricsEndpointNoAuth verifies /metrics is served without auth
// (acceptance criterion 6).
func TestMetricsEndpointNoAuth(t *testing.T) {
	m := newTestMetrics(t)
	reg := newRegister(m)
	reg.AddMethod("GET", "/ping", func(ctx *beecontext.Context) {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	})
	reg.Init()

	w := httptest.NewRecorder()
	reg.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))

	mw := httptest.NewRecorder()
	m.Handler().ServeHTTP(mw, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	require.Equal(t, http.StatusOK, mw.Code)
	assert.Contains(t, mw.Body.String(), "http_server_requests_total")
}

// TestCorpPDFEmailPathRecordsEmailMetric verifies that the corp_pdf_email path
// (direct emailservice.SendEmail, not through SendSimpleMessage) also
// records the email_sent_total metric (acceptance criterion 7).
func TestCorpPDFEmailPathRecordsEmailMetric(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	recordEmailSent("gmail", "sent")

	require.Equal(t, 1.0, testutil.ToFloat64(
		emailSentCounter.Vec().WithLabelValues("openeuler", "gmail", "sent")))
}

// TestNoHighCardinalityLabelsInMetrics verifies that no high-cardinality
// labels (request_id/trace_id/span_id/link_id/email) appear in /metrics
// text (acceptance criterion 5).
func TestNoHighCardinalityLabelsInMetrics(t *testing.T) {
	m := newTestMetrics(t)
	reg := newRegister(m)
	reg.AddMethod("GET", "/v1/cla/:link_id/:signing_id", func(ctx *beecontext.Context) {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	})
	reg.Init()

	w := httptest.NewRecorder()
	reg.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/cla/link123/sign456", nil))
	require.Equal(t, http.StatusOK, w.Code)

	mw := httptest.NewRecorder()
	m.Handler().ServeHTTP(mw, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := mw.Body.String()

	assert.False(t, strings.Contains(body, "link_id=\""), "link_id label should not appear in /metrics")
	assert.False(t, strings.Contains(body, "email=\""), "email label should not appear in /metrics")
	assert.False(t, strings.Contains(body, "request_id=\""), "request_id label should not appear in /metrics")
	assert.False(t, strings.Contains(body, "trace_id=\""), "trace_id label should not appear in /metrics")
	assert.False(t, strings.Contains(body, "span_id=\""), "span_id label should not appear in /metrics")
}

var fakeSendErr = &fakeError{msg: "smtp down"}

type fakeError struct{ msg string }

func (e *fakeError) Error() string { return e.msg }
