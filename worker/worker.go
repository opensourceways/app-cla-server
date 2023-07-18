package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
)

var (
	worker       IEmailWorker
	pdfGenerator pdf.IPDFGenerator
)

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

func Init(g pdf.IPDFGenerator) {
	pdfGenerator = g

	worker = &emailWorker{
		stop: make(chan struct{}),
	}
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

		w.tryToSendEmail(action)
	}

	w.wg.Add(1)
	go f(emailPlatform, *msg)
}

func (w *emailWorker) tryToSendEmail(action func() error) {
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
			break
		}
		logs.Error(err)

		reset(i > 0) // timer must be expired when i > 0

		select {
		case <-w.stop:
			return
		case <-t.C:
		}
	}
}
