package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
	"github.com/opensourceways/obs-sdk/go/metrics"
)

var (
	worker       IEmailWorker
	pdfGenerator pdf.IPDFGenerator

	// 业务计数器（邮件/PDF），由 main.initObservability 注入；nil 时打点为空操作。
	emailCounter *metrics.CounterVec
	pdfCounter   *metrics.CounterVec
)

// SetBusinessMetrics 注入业务计数器。nil 表示不采集对应指标（保持原审计日志行为）。
func SetBusinessMetrics(emailSent, pdfGen *metrics.CounterVec) {
	emailCounter = emailSent
	pdfCounter = pdfGen
}

// incEmail 在部署级默认 community 上记一条邮件结果计数；计数器未注入时为空操作。
func incEmail(platform, outcome string) {
	if emailCounter != nil {
		emailCounter.Inc(platform, outcome)
	}
}

// incPDF 在部署级默认 community 上记一条 PDF 生成结果计数；计数器未注入时为空操作。
func incPDF(outcome string) {
	if pdfCounter != nil {
		pdfCounter.Inc(outcome)
	}
}

type EmailMessage = emailservice.EmailMessage

type IEmailWorker interface {
	GenCLAPDFForCorporationAndSendIt(
		string, *models.OrgInfo, *models.CLAInfo, *models.CorporationSigning,
	)
	SendSimpleMessage(platform string, msg *EmailMessage)
	Shutdown()
}

func GetEmailWorker() IEmailWorker {
	return worker
}

// SetEmailWorker 注入一个 email worker 实例（主要用于测试中替换为假 worker），
// 返回还原函数以恢复原值。调用方应在测试结束后调用还原函数。
func SetEmailWorker(w IEmailWorker) func() {
	prev := worker
	worker = w
	return func() { worker = prev }
}

func Init(g pdf.IPDFGenerator) {
	pdfGenerator = g

	worker = &emailWorker{
		stop: make(chan struct{}),
	}
}

// SetPDFGenerator 注入一个 PDF 生成器实例（主要用于测试替换为假实现），
// 返回还原函数以恢复原值。调用方应在测试结束后调用还原函数。
func SetPDFGenerator(g pdf.IPDFGenerator) func() {
	prev := pdfGenerator
	pdfGenerator = g
	return func() { pdfGenerator = prev }
}

func Exit() {
	if worker != nil {
		worker.Shutdown()
	}
}

type emailWorker struct {
	wg   sync.WaitGroup
	stop chan struct{}
}

func (w *emailWorker) Shutdown() {
	close(w.stop)

	logs.Info("worker exit")

	// Handle remaining requests
	w.wg.Wait()
}

func (w *emailWorker) GenCLAPDFForCorporationAndSendIt(
	linkID string,
	orgInfo *models.OrgInfo,
	claInfo *models.CLAInfo,
	signing *models.CorporationSigning,
) {
	f := func(impl *corpPDFEmail) {
		w.tryToSendEmail(func() error {
			err := impl.do()
			if err != nil {
				err = fmt.Errorf(
					"send corp pdf of link:%s, %s", impl.linkID, err.Error(),
				)
			}

			return err
		})

		impl.clean()

		w.wg.Done()
	}

	w.wg.Add(1)

	go f(newCorpPDFEmail(linkID, orgInfo, claInfo, signing))
}

func (w *emailWorker) SendSimpleMessage(emailPlatform string, msg *EmailMessage) {
	f := func(platform string, msg1 EmailMessage) {
		defer func() {
			if msg1.HasSecret {
				msg1.ClearContent()
			}

			w.wg.Done()
		}()

		action := func() error {
			if err := emailservice.SendEmail(platform, &msg1); err != nil {
				return fmt.Errorf("error to send email, err:%s", err.Error())
			}
			return nil
		}

		// [audit] 据发信最终结果记录审计日志（成功 Info / 失败 Error），覆盖所有走 SendSimpleMessage 的邮件。
		if w.tryToSendEmail(action) {
			logs.Info("[audit] email_result: outcome=sent, platform=%s, to=%v, subject=%s", platform, msg1.To, msg1.Subject)
			incEmail(platform, "sent")
		} else {
			logs.Error("[audit] email_result: outcome=failed, platform=%s, to=%v, subject=%s", platform, msg1.To, msg1.Subject)
			incEmail(platform, "failed")
		}
	}

	w.wg.Add(1)
	go f(emailPlatform, *msg)
}

// tryToSendEmail 执行带重试的发信动作，返回是否最终发送成功（true=成功，false=重试耗尽或被停止）。
func (w *emailWorker) tryToSendEmail(action func() error) bool {
	t := time.NewTimer(1 * time.Minute)
	defer t.Stop()

	reset := func(expired bool) {
		if !expired && !t.Stop() {
			<-t.C
		}
		t.Reset(1 * time.Minute)
	}

	for i := 0; i < 10; i++ {
		err := action()
		if err == nil {
			return true
		}

		logs.Error(err)

		reset(i > 0) // timer must be expired when i > 0

		select {
		case <-w.stop:
			return false
		case <-t.C:
		}
	}

	return false
}
