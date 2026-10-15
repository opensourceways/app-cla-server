package worker

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
	beecontext "github.com/beego/beego/v2/server/web/context"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
	obsmetrics "github.com/opensourceways/obs-sdk/go/metrics"
	"github.com/opensourceways/obs-sdk/go/middleware/beegomw"
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

// TestEmailSentMetricValue verifies the counter value via registry gather.
func TestEmailSentMetricValue(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	recordEmailSent("gmail", "sent")
	recordEmailSent("gmail", "sent")
	recordEmailSent("gmail", "failed")

	require.Equal(t, 2.0, findCounterValue(t, m, "cla_email_sent_total", map[string]string{
		"community": "openeuler", "platform": "gmail", "outcome": "sent",
	}))
	require.Equal(t, 1.0, findCounterValue(t, m, "cla_email_sent_total", map[string]string{
		"community": "openeuler", "platform": "gmail", "outcome": "failed",
	}))
}

// TestPDFGeneratedMetricValue verifies the pdf counter value via registry gather.
func TestPDFGeneratedMetricValue(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	recordPDFGenerated("success")
	recordPDFGenerated("success")
	recordPDFGenerated("failed")

	require.Equal(t, 2.0, findCounterValue(t, m, "cla_pdf_generated_total", map[string]string{
		"community": "openeuler", "outcome": "success",
	}))
	require.Equal(t, 1.0, findCounterValue(t, m, "cla_pdf_generated_total", map[string]string{
		"community": "openeuler", "outcome": "failed",
	}))
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

	require.Equal(t, 1.0, findCounterValue(t, m, "cla_email_sent_total", map[string]string{
		"community": "openeuler", "platform": "test-metric-success", "outcome": "sent",
	}))
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

	require.GreaterOrEqual(t, findCounterValue(t, m, "cla_email_sent_total", map[string]string{
		"community": "openeuler", "platform": "test-metric-fail", "outcome": "failed",
	}), 1.0)
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

// --- corp_pdf_email.go integration tests ---
// These tests exercise the real do()/genFile() code paths (acceptance criterion 7)
// by injecting a fake pdf.IPDFGenerator and a fake emailservice, then asserting
// on real /metrics text output.

// fakePDFGenerator implements pdf.IPDFGenerator for testing.
type fakePDFGenerator struct {
	filePath string
	err      error
	calls    int
}

func (f *fakePDFGenerator) GenPDFForCorporationSigning(
	linkID, claFile string,
	signing *models.CorporationSigning,
	claFields []models.CLAField,
) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.filePath, nil
}

func setPDFGenerator(g pdf.IPDFGenerator) func() {
	prev := pdfGenerator
	pdfGenerator = g
	return func() { pdfGenerator = prev }
}

func createTempPDFFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "test-*.pdf")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func newTestCorpPDFEmail(platform string) *corpPDFEmail {
	return &corpPDFEmail{
		linkID: "test-link-id",
		msg:    &EmailMessage{To: []string{"admin@example.com"}, Subject: "test"},
		orgInfo: models.OrgInfo{
			OrgEmailPlatform: platform,
			OrgAlias:         "test-org",
			OrgEmail:         "noreply@example.com",
		},
		claInfo: models.CLAInfo{
			CLAFile: "test-cla.pdf",
			Fields:  []models.CLAField{{ID: "corp", Title: "Corporation"}},
		},
		signing: models.CorporationSigning{
			CorporationSigningBasicInfo: models.CorporationSigningBasicInfo{
				AdminName:  "admin",
				AdminEmail: "admin@example.com",
				Date:       "2024-01-01",
			},
			Info: models.TypeSigningInfo{"corp": "Test Corp"},
		},
	}
}

func gatherMetricsText(t *testing.T, m *obsmetrics.Metrics) string {
	t.Helper()
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	return w.Body.String()
}

// TestCorpPDFEmailDoRecordsMetrics verifies that corp_pdf_email.do() records
// both email_sent_total{outcome="sent"} and pdf_generated_total{outcome="success"}
// through the real code path (acceptance criterion 7).
func TestCorpPDFEmailDoRecordsMetrics(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	platform := "test-pdf-email-ok"
	fake := &fakeEmailService{}
	emailservice.Register(platform, fake)

	tmpFile := createTempPDFFile(t)
	t.Cleanup(func() { os.Remove(tmpFile) })

	restore := setPDFGenerator(&fakePDFGenerator{filePath: tmpFile})
	t.Cleanup(restore)

	impl := newTestCorpPDFEmail(platform)
	err := impl.do()
	require.NoError(t, err)

	body := gatherMetricsText(t, m)
	assert.Contains(t, body, "cla_email_sent_total")
	assert.Contains(t, body, "cla_pdf_generated_total")
	assert.Contains(t, body, `platform="test-pdf-email-ok"`)
	assert.Contains(t, body, `outcome="sent"`)
	assert.Contains(t, body, `outcome="success"`)

	require.Equal(t, 1.0, findCounterValue(t, m, "cla_email_sent_total", map[string]string{
		"community": "openeuler", "platform": platform, "outcome": "sent",
	}))
	require.Equal(t, 1.0, findCounterValue(t, m, "cla_pdf_generated_total", map[string]string{
		"community": "openeuler", "outcome": "success",
	}))
	require.Equal(t, 1, fake.calls)
}

// TestCorpPDFEmailDoEmailFailureRecordsFailedMetric verifies that when
// emailservice.SendEmail fails inside do(), the email_sent_total{outcome="failed"}
// metric is recorded (acceptance criterion 7).
func TestCorpPDFEmailDoEmailFailureRecordsFailedMetric(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	platform := "test-pdf-email-fail"
	fake := &fakeEmailService{err: fakeSendErr}
	emailservice.Register(platform, fake)

	tmpFile := createTempPDFFile(t)
	t.Cleanup(func() { os.Remove(tmpFile) })

	restore := setPDFGenerator(&fakePDFGenerator{filePath: tmpFile})
	t.Cleanup(restore)

	impl := newTestCorpPDFEmail(platform)
	err := impl.do()
	require.Error(t, err)

	body := gatherMetricsText(t, m)
	assert.Contains(t, body, `outcome="failed"`)
	assert.Contains(t, body, `outcome="success"`)

	require.Equal(t, 1.0, findCounterValue(t, m, "cla_email_sent_total", map[string]string{
		"community": "openeuler", "platform": platform, "outcome": "failed",
	}))
	require.Equal(t, 1.0, findCounterValue(t, m, "cla_pdf_generated_total", map[string]string{
		"community": "openeuler", "outcome": "success",
	}))
}

// TestCorpPDFEmailDoPDFFailureRecordsFailedMetric verifies that when
// pdfGenerator fails inside genFile(), the pdf_generated_total{outcome="failed"}
// metric is recorded and do() returns early (acceptance criterion 7).
func TestCorpPDFEmailDoPDFFailureRecordsFailedMetric(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	platform := "test-pdf-gen-fail"
	emailservice.Register(platform, &fakeEmailService{})

	restore := setPDFGenerator(&fakePDFGenerator{err: fakeSendErr})
	t.Cleanup(restore)

	impl := newTestCorpPDFEmail(platform)
	err := impl.do()
	require.Error(t, err)

	body := gatherMetricsText(t, m)
	assert.Contains(t, body, "cla_pdf_generated_total")
	assert.Contains(t, body, `outcome="failed"`)

	require.Equal(t, 1.0, findCounterValue(t, m, "cla_pdf_generated_total", map[string]string{
		"community": "openeuler", "outcome": "failed",
	}))
}

// TestCorpPDFEmailDoGenMsgFailure verifies that when genMsg() fails inside
// do() (no email templates loaded), the pdf_generated_total{outcome="success"}
// metric is still recorded and do() returns the error.
func TestCorpPDFEmailDoGenMsgFailure(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	platform := "test-pdf-genmsg-fail"
	emailservice.Register(platform, &fakeEmailService{})

	tmpFile := createTempPDFFile(t)
	t.Cleanup(func() { os.Remove(tmpFile) })

	restore := setPDFGenerator(&fakePDFGenerator{filePath: tmpFile})
	t.Cleanup(restore)

	impl := newTestCorpPDFEmail(platform)
	impl.msg = nil

	err := impl.do()
	require.Error(t, err)

	require.Equal(t, 1.0, findCounterValue(t, m, "cla_pdf_generated_total", map[string]string{
		"community": "openeuler", "outcome": "success",
	}))
}

// TestCorpPDFEmailGenFileSkipsWhenFileExists verifies that genFile() returns
// early without calling pdfGenerator when the PDF file already exists,
// so no pdf_generated_total metric is recorded.
func TestCorpPDFEmailGenFileSkipsWhenFileExists(t *testing.T) {
	resetBusinessMetrics()
	m := newTestMetrics(t)
	InitBusinessMetrics(m)

	tmpFile := createTempPDFFile(t)
	t.Cleanup(func() { os.Remove(tmpFile) })

	fg := &fakePDFGenerator{filePath: tmpFile}
	restore := setPDFGenerator(fg)
	t.Cleanup(restore)

	impl := newTestCorpPDFEmail("test-skip")
	impl.pdfFilePath = tmpFile

	err := impl.genFile()
	require.NoError(t, err)
	require.Equal(t, 0, fg.calls)
}

// TestCorpPDFEmailCleanRemovesFile verifies that clean() removes the PDF file
// when it exists.
func TestCorpPDFEmailCleanRemovesFile(t *testing.T) {
	tmpFile := createTempPDFFile(t)

	impl := &corpPDFEmail{pdfFilePath: tmpFile}
	impl.clean()

	_, err := os.Stat(tmpFile)
	require.True(t, os.IsNotExist(err), "temp file should be removed")
}

// TestCorpPDFEmailCleanNoFile verifies that clean() does not panic when no
// PDF file exists.
func TestCorpPDFEmailCleanNoFile(t *testing.T) {
	impl := &corpPDFEmail{}
	impl.clean()
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
