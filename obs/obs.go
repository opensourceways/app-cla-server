// Package obs 把 obs-sdk-go 的观测能力（HTTP 服务器指标 + 业务计数器）
// 装配到 app-cla-server。所有指标挂在同一个 *metrics.Metrics 的独立 registry 上，
// 不与 prometheus 全局 DefaultRegisterer 冲突。
//
// 接入方式（main.go）：
//
//	m := obs.NewMetrics()
//	beego.InsertFilterChain("/*", obs.FilterChain(m))
//	beego.Handler("/metrics", obs.MetricsHandler(m))
//	bm := obs.NewBusinessMetrics(m)
//	worker.SetBusinessMetrics(bm.EmailSent, bm.PDFGen)
//
// 注意：beegomw.Middleware 在调用时刻注册指标，每个 registry 只能调用一次
// （重复注册同名指标会 panic）。因此生产代码与每个测试用例都各自构造独立的
// *metrics.Metrics，且只调用一次 FilterChain / NewBusinessMetrics。
package obs

import (
	"net/http"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/opensourceways/obs-sdk/go/metrics"
	"github.com/opensourceways/obs-sdk/go/middleware/beegomw"
)

// ServiceName 是指标的 service label 取值，与 deploy/app.conf 的 appname 对齐。
const ServiceName = "app-cla-server"

// NewMetrics 创建本服务的指标装配实例。env / instance / community 由 SDK 按
// OBS_* 环境变量解析（部署侧注入），此处只钉 service 名。
func NewMetrics() *metrics.Metrics {
	return metrics.New(metrics.Config{Service: ServiceName})
}

// FilterChain 返回挂在给定 metrics 上的 beego 过滤器链。
// path label 取 beego 命中路由后写入的 RouterPattern（路由模板），
// 未命中路由统一记常量 unmatched，保证 label 基数有界。
// 该过滤器包住 serveHttp 本身，因此 Controller.StopRun()（鉴权失败 401/403）
// 也会被恰好收尾一次，不会被静默漏掉。
func FilterChain(m *metrics.Metrics) beego.FilterChain {
	return beegomw.Middleware(beegomw.Options{Metrics: m})
}

// MetricsHandler 返回 /metrics 端点的 HTTP handler（Prometheus text exposition）。
// 该端点以 beego.Handler 注册为裸 http.Handler，不走任何 controller 的 Prepare
// 鉴权，Prometheus 可直接抓取。
func MetricsHandler(m *metrics.Metrics) http.Handler {
	return m.Handler()
}

// BusinessMetrics 汇总本服务的业务计数器。label 不含高基数项
// （邮箱、subject、link_id 仅入日志，不入 label）。
type BusinessMetrics struct {
	// EmailSent: cla_email_sent_total{platform, outcome}，outcome=sent/failed。
	EmailSent *metrics.CounterVec
	// PDFGen: cla_pdf_generated_total{outcome}，outcome=success/failed。
	PDFGen *metrics.CounterVec
}

// NewBusinessMetrics 在给定 metrics 上注册业务计数器。
// 与 FilterChain 注册的 HTTP 指标共享同一 registry，统一经 /metrics 暴露。
func NewBusinessMetrics(m *metrics.Metrics) *BusinessMetrics {
	return &BusinessMetrics{
		EmailSent: m.NewCounterVec(
			"cla_email_sent_total",
			"Number of CLA notification emails sent",
			"platform", "outcome",
		),
		PDFGen: m.NewCounterVec(
			"cla_pdf_generated_total",
			"Number of corporation CLA PDFs generated",
			"outcome",
		),
	}
}
