package obs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
	beecontext "github.com/beego/beego/v2/server/web/context"
	"github.com/opensourceways/obs-sdk/go/metrics"
)

// TestMain 把 OBS_* 环境变量钉成确定值，避免用例依赖 runner 的 hostname / 环境，
// 也使 community / env / instance const label 取值确定，便于断言。
// service 恒由 obs.NewMetrics 钉为 app-cla-server。
func TestMain(m *testing.M) {
	os.Setenv("OBS_ENV", "test")
	os.Setenv("OBS_INSTANCE", "test-pod")
	os.Setenv("OBS_COMMUNITY", "test-community")
	os.Exit(m.Run())
}

// metricsText 抓取一次 /metrics 输出文本。断言一律落在这段真实 Prometheus text
// exposition 上，而不是中间变量（完成标准 #7）。
func metricsText(t *testing.T, m *metrics.Metrics) string {
	t.Helper()
	rec := httptest.NewRecorder()
	MetricsHandler(m).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/metrics returned %d, want 200", rec.Code)
	}
	return rec.Body.String()
}

// newRegister 构造一个独立的 ControllerRegister 并挂上本服务的 obs 过滤器链。
// 用独立注册器而非全局 BeeApp，避免用例间共享路由/过滤器链，也避免重复注册指标。
func newRegister(t *testing.T, m *metrics.Metrics) *beego.ControllerRegister {
	t.Helper()
	reg := beego.NewControllerRegister()
	reg.InsertFilterChain("/*", FilterChain(m))
	return reg
}

// stopRunController 复刻 controllers/access_controller.go 的鉴权失败出口：
// 先写响应状态码，再 StopRun()。StopRun 内部 panic(ErrAbort)，由 serveHttp 自己的
// RecoverFunc 就地吞掉，AfterExec / FinishRouter 过滤器被整个跳过 —— 用
// InsertFilterChain 包住 serveHttp 才能保证这类请求恰好收尾一次（完成标准 #2）。
type stopRunController struct {
	beego.Controller
}

func (c *stopRunController) Post() {
	c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
	c.StopRun()
}

func (c *stopRunController) Get() {
	c.Ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
	c.StopRun()
}

// TestMetricsHandlerAccessibleNoAuth 覆盖完成标准 #6 的 handler 层：
// /metrics 无需任何鉴权头即可抓到、返回 200，且常驻 label（service/community）非空。
// const label 只挂在已注册并产生过 series 的指标上，故此处先记一条业务计数。
func TestMetricsHandlerAccessibleNoAuth(t *testing.T) {
	m := NewMetrics()
	bm := NewBusinessMetrics(m)
	bm.PDFGen.Inc("success")

	body := metricsText(t, m)

	for _, want := range []string{
		`service="app-cla-server"`,
		`env="test"`,
		`instance="test-pod"`,
		`community="test-community"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics text missing const label %s", want)
		}
	}
}

// TestHTTPServerMetricsRecorded 覆盖完成标准 #1：一次正常 200 请求后，
// /metrics 同时暴露 counter（http_server_requests_total）与 histogram
//（http_server_request_duration_seconds），series 带 method/path/status_code
// 以及 service/env/instance/community 四个常驻 label。
func TestHTTPServerMetricsRecorded(t *testing.T) {
	m := NewMetrics()
	reg := newRegister(t, m)
	reg.AddMethod("GET", "/ping", func(ctx *beecontext.Context) {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	})
	reg.Init()

	w := httptest.NewRecorder()
	reg.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET /ping: got %d, want 200", w.Code)
	}

	body := metricsText(t, m)
	for _, want := range []string{
		"# TYPE http_server_requests_total counter",
		`http_server_requests_total{`,
		`method="GET"`,
		`path="/ping"`,
		`status_code="200"`,
		"# TYPE http_server_request_duration_seconds histogram",
		`http_server_request_duration_seconds_bucket{`,
		`http_server_request_duration_seconds_count{`,
		`service="app-cla-server"`,
		`env="test"`,
		`instance="test-pod"`,
		`community="test-community"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics text missing %q", want)
		}
	}
}

// TestStopRunFailureStillRecorded 覆盖完成标准 #2：鉴权失败路径（先写状态码再 StopRun）
// 必须被计数。POST→401 对应无效 token，GET→403 对应权限不符，均经 apiPrepareWithAC
// 的 StopRun 出口。
func TestStopRunFailureStillRecorded(t *testing.T) {
	m := NewMetrics()
	reg := newRegister(t, m)
	reg.Add("/v1/link/:link_id", &stopRunController{})
	reg.Init()

	// 401：无效 token 的 StopRun 出口
	w := httptest.NewRecorder()
	reg.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/link/abc123", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("POST /v1/link/abc123: got %d, want 401", w.Code)
	}

	body := metricsText(t, m)
	if !strings.Contains(body, `http_server_requests_total{`) {
		t.Fatal("http_server_requests_total missing from /metrics")
	}
	if !strings.Contains(body, `status_code="401"`) {
		t.Fatal("401 StopRun request not recorded: status_code=\"401\" missing")
	}
	if !strings.Contains(body, `path="/v1/link/:link_id"`) {
		t.Fatal("path label is not the route template /v1/link/:link_id")
	}
	if strings.Contains(body, "abc123") {
		t.Fatal("raw link_id value leaked into /metrics")
	}

	// 403：权限不符的 StopRun 出口
	w2 := httptest.NewRecorder()
	reg.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/v1/link/def456", nil))
	if w2.Code != http.StatusForbidden {
		t.Fatalf("GET /v1/link/def456: got %d, want 403", w2.Code)
	}
	if !strings.Contains(metricsText(t, m), `status_code="403"`) {
		t.Fatal("403 StopRun request not recorded: status_code=\"403\" missing")
	}
}

// TestPathLabelIsRouteTemplate 覆盖完成标准 #3：path label 必须是路由模板，
// 负向断言真实 link_id 值不出现在 /metrics 文本里。
func TestPathLabelIsRouteTemplate(t *testing.T) {
	m := NewMetrics()
	reg := newRegister(t, m)
	reg.AddMethod("GET", "/v1/cla/:link_id", func(ctx *beecontext.Context) {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	})
	reg.Init()

	w := httptest.NewRecorder()
	reg.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/cla/a1b2c3d4", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET /v1/cla/a1b2c3d4: got %d, want 200", w.Code)
	}

	body := metricsText(t, m)
	if !strings.Contains(body, `path="/v1/cla/:link_id"`) {
		t.Fatal("path label is not the route template /v1/cla/:link_id")
	}
	if strings.Contains(body, "a1b2c3d4") {
		t.Fatal("raw link_id value leaked into /metrics")
	}
}

// TestUnmatchedRouteBounded 覆盖完成标准 #4：未命中路由只产生一条 path=\"unmatched\"
// series，打两个不同随机路径验证基数有界，且随机路径不出现在 /metrics 文本里。
func TestUnmatchedRouteBounded(t *testing.T) {
	m := NewMetrics()
	reg := newRegister(t, m)
	reg.AddMethod("GET", "/ping", func(ctx *beecontext.Context) {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	})
	reg.Init()

	for _, p := range []string{"/nope/11111111", "/nope/22222222"} {
		reg.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, p, nil))
	}

	body := metricsText(t, m)
	if !strings.Contains(body, `path="unmatched"`) {
		t.Fatal("unmatched path label missing")
	}
	if strings.Contains(body, "/nope/11111111") || strings.Contains(body, "/nope/22222222") {
		t.Fatal("random unmatched path leaked into /metrics")
	}
}

// TestNoHighCardinalityLabels 覆盖完成标准 #5：request_id / trace_id / span_id /
// link_id / email 不得作为高基数 label 出现。带入境 X-Request-Id 的值也不得入文本。
func TestNoHighCardinalityLabels(t *testing.T) {
	m := NewMetrics()
	reg := newRegister(t, m)
	reg.AddMethod("GET", "/v1/cla/:link_id", func(ctx *beecontext.Context) {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	})
	reg.Init()

	req := httptest.NewRequest(http.MethodGet, "/v1/cla/a1b2c3d4", nil)
	req.Header.Set("X-Request-Id", "secret-req-id-123")
	reg.ServeHTTP(httptest.NewRecorder(), req)

	body := metricsText(t, m)
	// 高基数 label 名不得出现（label 语法 name=）。
	for _, bad := range []string{
		`request_id=`,
		`trace_id=`,
		`span_id=`,
		`link_id=`,
		`email=`,
	} {
		if strings.Contains(body, bad) {
			t.Errorf("/metrics text contains high-cardinality label %q", bad)
		}
	}
	// 入境 request_id 的值与原始 link_id 值不得出现。
	for _, bad := range []string{"secret-req-id-123", "a1b2c3d4"} {
		if strings.Contains(body, bad) {
			t.Errorf("/metrics text leaked value %q", bad)
		}
	}
}

// TestMetricsEndpointReachableWithoutAuth 覆盖完成标准 #6：/metrics 端点通过框架路由
//（经 /* 过滤器链）即可抓到，无需任何鉴权头。先打一次 /ping 产生 series，
// 再抓 /metrics 验证正文含 http_server_requests_total。
func TestMetricsEndpointReachableWithoutAuth(t *testing.T) {
	m := NewMetrics()
	reg := newRegister(t, m)
	reg.AddMethod("GET", "/ping", func(ctx *beecontext.Context) {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	})
	reg.Handler("/metrics", MetricsHandler(m))
	reg.Init()

	// 先打一次 /ping 产生一条 http_server_requests_total series。
	reg.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ping", nil))

	// /metrics 无需鉴权头即可抓到。
	w := httptest.NewRecorder()
	reg.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("/metrics through router: got %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "http_server_requests_total") {
		t.Fatal("/metrics body missing http_server_requests_total")
	}
}

// TestBusinessMetricsExposed 覆盖第五节业务指标：cla_email_sent_total{platform,outcome}
// 与 cla_pdf_generated_total{outcome} 经 /metrics 暴露，且与 HTTP 指标共享同一 registry。
func TestBusinessMetricsExposed(t *testing.T) {
	m := NewMetrics()
	bm := NewBusinessMetrics(m)

	bm.EmailSent.Inc("github", "sent")
	bm.EmailSent.Inc("gitee", "failed")
	bm.PDFGen.Inc("success")
	bm.PDFGen.Inc("failed")

	body := metricsText(t, m)
	for _, want := range []string{
		"# TYPE cla_email_sent_total counter",
		`cla_email_sent_total{`,
		`platform="github"`,
		`outcome="sent"`,
		`platform="gitee"`,
		`outcome="failed"`,
		"# TYPE cla_pdf_generated_total counter",
		`cla_pdf_generated_total{`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics text missing %q", want)
		}
	}
	// 高基数项（to/subject/link_id）不得作为 label 出现。
	for _, bad := range []string{`to=`, `subject=`, `link_id=`} {
		if strings.Contains(body, bad) {
			t.Errorf("/metrics text contains high-cardinality business label %q", bad)
		}
	}
}
