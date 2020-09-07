package worker

import (
	"os"
	"sync"
	"time"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/email"
	"github.com/zengchen1024/cla-server/models"
	"github.com/zengchen1024/cla-server/pdf"
	"github.com/zengchen1024/cla-server/util"
)

var worker IEmailWorker

type IEmailWorker interface {
	GenCLAPDFForCorporationAndSendIt(claOrg *models.CLAOrg, signing *models.CorporationSigning, cla *models.CLA, emailCfg *models.OrgEmail)
}

func GetEmailWorker() IEmailWorker {
	return worker
}

func InitEmailWorker(g pdf.IPDFGenerator) {
	worker = &emailWorker{pdfGenerator: g}
}

type emailWorker struct {
	pdfGenerator pdf.IPDFGenerator
	wg           sync.WaitGroup
	shutdown     bool
}

func (this *emailWorker) Shutdown() {
	this.shutdown = true

	// Handle remaining requests
	this.wg.Wait()
}

func (this *emailWorker) GenCLAPDFForCorporationAndSendIt(claOrg *models.CLAOrg, signing *models.CorporationSigning, cla *models.CLA, emailCfg *models.OrgEmail) {
	f := func() {
		defer func() {
			this.wg.Done()
		}()

		wait := func() { time.Sleep(time.Minute * time.Duration(1)) }

		file := ""
		for {
			if this.shutdown {
				beego.Info("exit email worker forcedly")
				break
			}

			if file == "" || util.IsFileNotExist(file) {
				file1, err := this.pdfGenerator.GenCLAPDFForCorporation(claOrg, signing, cla)
				if err != nil {
					beego.Info(err)
					wait()
					continue
				}
				file = file1
			}

			e, err := email.GetEmailClient(emailCfg.Platform)
			if err != nil {
				beego.Info(err)
				wait()
				continue
			}

			msg := email.EmailMessage{
				To:         signing.AdminEmail,
				Subject:    "pdf signing",
				Content:    "pdf",
				Attachment: file,
			}
			if err := e.SendEmail(*emailCfg.Token, msg); err != nil {
				beego.Info(err)
				wait()
				continue
			}

			os.Remove(file)
			break
		}
	}

	this.wg.Add(1)
	go f()
}
