package worker

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/obs"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
	"github.com/opensourceways/obs-sdk/go/metrics"
)

// readMetrics 抓取一次 /metrics 输出文本，断言落在真实 Prometheus text exposition 上。
func readMetrics(t *testing.T, m *metrics.Metrics) string {
	t.Helper()
	rec := httptest.NewRecorder()
	obs.MetricsHandler(m).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/metrics returned %d, want 200", rec.Code)
	}
	return rec.Body.String()
}

// fakePDFGenerator 替换真实 PDF 生成器，按预设返回路径或错误。
type fakePDFGenerator struct {
	path string
	err  error
}

func (f *fakePDFGenerator) GenPDFForCorporationSigning(
	linkID, claFile string, signing *models.CorporationSigning,
	claFields []models.CLAField,
) (string, error) {
	return f.path, f.err
}

// TestSendSimpleMessageIncrementsEmailCounterOnSuccess 验证发信成功时
// cla_email_sent_total{platform, outcome="sent"} 计数 +1，并出现在 /metrics 文本里。
func TestSendSimpleMessageIncrementsEmailCounterOnSuccess(t *testing.T) {
	m := obs.NewMetrics()
	bm := obs.NewBusinessMetrics(m)
	SetBusinessMetrics(bm.EmailSent, bm.PDFGen)
	defer SetBusinessMetrics(nil, nil)

	fake := &fakeEmailService{}
	emailservice.Register("plat-ctr-ok", fake)

	w := &emailWorker{stop: make(chan struct{})}
	w.SendSimpleMessage("plat-ctr-ok", &EmailMessage{To: []string{"a@example.com"}, Subject: "hi"})
	w.Shutdown()

	body := readMetrics(t, m)
	for _, want := range []string{
		`cla_email_sent_total{`,
		`platform="plat-ctr-ok"`,
		`outcome="sent"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics text missing %q", want)
		}
	}
	if strings.Contains(body, "a@example.com") {
		t.Error("recipient email address leaked into /metrics")
	}
}

// TestSendSimpleMessageIncrementsEmailCounterOnFailure 通过总是失败的假 emailservice
// 再由 Shutdown 关闭 stop 使重试快速终止，验证失败分支记 outcome="failed"。
func TestSendSimpleMessageIncrementsEmailCounterOnFailure(t *testing.T) {
	m := obs.NewMetrics()
	bm := obs.NewBusinessMetrics(m)
	SetBusinessMetrics(bm.EmailSent, bm.PDFGen)
	defer SetBusinessMetrics(nil, nil)

	fake := &fakeEmailService{err: errors.New("smtp down")}
	emailservice.Register("plat-ctr-fail", fake)

	w := &emailWorker{stop: make(chan struct{})}
	w.SendSimpleMessage("plat-ctr-fail", &EmailMessage{To: []string{"a@example.com"}, Subject: "hi"})
	w.Shutdown() // 关闭 stop -> 首次失败后经 stop 分支快速返回 false

	body := readMetrics(t, m)
	if !strings.Contains(body, `cla_email_sent_total{`) {
		t.Fatal("cla_email_sent_total missing from /metrics")
	}
	if !strings.Contains(body, `platform="plat-ctr-fail"`) {
		t.Fatal("platform label missing")
	}
	if !strings.Contains(body, `outcome="failed"`) {
		t.Fatal("failed outcome not recorded")
	}
}

// TestGenFileIncrementsPDFCounterOnSuccess 验证 PDF 生成成功时
// cla_pdf_generated_total{outcome="success"} 计数 +1。
func TestGenFileIncrementsPDFCounterOnSuccess(t *testing.T) {
	m := obs.NewMetrics()
	bm := obs.NewBusinessMetrics(m)
	SetBusinessMetrics(bm.EmailSent, bm.PDFGen)
	defer SetBusinessMetrics(nil, nil)

	restore := SetPDFGenerator(&fakePDFGenerator{path: "/tmp/cla-test.pdf"})
	defer restore()

	impl := newCorpPDFEmail(
		"link1", &models.OrgInfo{}, &models.CLAInfo{}, &models.CorporationSigning{},
	)
	if err := impl.genFile(); err != nil {
		t.Fatalf("genFile: unexpected error: %v", err)
	}

	body := readMetrics(t, m)
	if !strings.Contains(body, `cla_pdf_generated_total{`) {
		t.Fatal("cla_pdf_generated_total missing from /metrics")
	}
	if !strings.Contains(body, `outcome="success"`) {
		t.Fatal("success outcome not recorded")
	}
}

// TestGenFileIncrementsPDFCounterOnFailure 验证 PDF 生成失败时
// cla_pdf_generated_total{outcome="failed"} 计数 +1。
func TestGenFileIncrementsPDFCounterOnFailure(t *testing.T) {
	m := obs.NewMetrics()
	bm := obs.NewBusinessMetrics(m)
	SetBusinessMetrics(bm.EmailSent, bm.PDFGen)
	defer SetBusinessMetrics(nil, nil)

	restore := SetPDFGenerator(&fakePDFGenerator{err: errors.New("pdf gen failed")})
	defer restore()

	impl := newCorpPDFEmail(
		"link1", &models.OrgInfo{}, &models.CLAInfo{}, &models.CorporationSigning{},
	)
	if err := impl.genFile(); err == nil {
		t.Fatal("genFile: expected error, got nil")
	}

	body := readMetrics(t, m)
	if !strings.Contains(body, `cla_pdf_generated_total{`) {
		t.Fatal("cla_pdf_generated_total missing from /metrics")
	}
	if !strings.Contains(body, `outcome="failed"`) {
		t.Fatal("failed outcome not recorded")
	}
}

// TestIncHelpersNoOpWithoutCounter 验证计数器未注入时（nil）打点为空操作，不 panic。
// 对应未注入指标的旧测试路径（保持原有行为）。
func TestIncHelpersNoOpWithoutCounter(t *testing.T) {
	SetBusinessMetrics(nil, nil)

	// 不应 panic。
	incEmail("github", "sent")
	incEmail("gitee", "failed")
	incPDF("success")
	incPDF("failed")
}
