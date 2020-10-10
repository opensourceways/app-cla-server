package worker

import (
	"os"
	"sync"
	"time"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

var worker IEmailWorker

type IEmailWorker interface {
	GenCLAPDFForCorporationAndSendIt(claOrg *models.CLAOrg, signing *models.CorporationSigning, cla *models.CLA, emailCfg *models.OrgEmail)
	SendSimpleMessage(emailCfg *models.OrgEmail, msg *email.EmailMessage)
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

		file := ""
		for {
			if this.shutdown {
				beego.Info("email worker exits forcedly")
				break
			}

			if file == "" || util.IsFileNotExist(file) {
				file1, err := this.pdfGenerator.GenCLAPDFForCorporation(claOrg, signing, cla)
				if err != nil {
					next(err)
					continue
				}
				file = file1
			}

			e, err := email.GetEmailClient(emailCfg.Platform)
			if err != nil {
				beego.Info(err)
				break
			}

			data := email.CorporationSigning{}
			msg, err := data.GenEmailMsg()
			if err != nil {
				next(err)
				continue
			}

			msg.To = []string{signing.AdminEmail}
			msg.Attachment = file

			if err := e.SendEmail(emailCfg.Token, msg); err != nil {
				next(err)
				continue
			}

			os.Remove(file)
			break
		}
	}

	this.wg.Add(1)
	go f()
}

func (this *emailWorker) SendSimpleMessage(emailCfg *models.OrgEmail, msg *email.EmailMessage) {
	f := func() {
		defer func() {
			this.wg.Done()
		}()

		for {
			if this.shutdown {
				beego.Info("email worker exits forcedly")
				break
			}

			e, err := email.GetEmailClient(emailCfg.Platform)
			if err != nil {
				beego.Info(err)
				break
			}

			if err := e.SendEmail(emailCfg.Token, msg); err != nil {
				next(err)
				continue
			}

			break
		}
	}

	this.wg.Add(1)
	go f()
}

func next(err error) {
	beego.Info(err)
	time.Sleep(time.Minute * time.Duration(1))

}
