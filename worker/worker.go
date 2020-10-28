package worker

import (
	"fmt"
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
	GenCLAPDFForCorporationAndSendIt(orgCLA *models.OrgCLA, signing *models.CorporationSigning, cla *models.CLA)
	SendSimpleMessage(orgEmail string, msg *email.EmailMessage)
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

func (this *emailWorker) GenCLAPDFForCorporationAndSendIt(orgCLA *models.OrgCLA, signing *models.CorporationSigning, cla *models.CLA) {
	f := func() {
		defer func() {
			this.wg.Done()
		}()

		emailCfg, ec, err := getEmailClient(orgCLA.OrgEmail)
		if err != nil {
			return
		}

		file := ""
		for {
			if this.shutdown {
				beego.Info("email worker exits forcedly")
				break
			}

			if file == "" || util.IsFileNotExist(file) {
				file1, err := this.pdfGenerator.GenPDFForCorporationSigning(orgCLA, signing, cla)
				if err != nil {
					next(fmt.Errorf(
						"Failed to generate pdf for corp signing(%s:%s:%s/%s): %s",
						orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID, util.EmailSuffix(signing.AdminEmail),
						err.Error()))
					continue
				}
				file = file1
			}

			data := email.CorporationSigning{}
			msg, err := data.GenEmailMsg()
			if err != nil {
				next(err)
				continue
			}
			msg.To = []string{signing.AdminEmail}
			msg.Attachment = file

			if err := ec.SendEmail(emailCfg.Token, msg); err != nil {
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

func (this *emailWorker) SendSimpleMessage(orgEmail string, msg *email.EmailMessage) {
	f := func() {
		defer func() {
			this.wg.Done()
		}()

		emailCfg, ec, err := getEmailClient(orgEmail)
		if err != nil {
			return
		}

		for {
			if this.shutdown {
				beego.Info("email worker exits forcedly")
				break
			}

			if err := ec.SendEmail(emailCfg.Token, msg); err != nil {
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
	beego.Info(err.Error())
	time.Sleep(time.Minute * time.Duration(1))

}

func getEmailClient(orgEmail string) (*models.OrgEmail, email.IEmail, error) {
	emailCfg := &models.OrgEmail{Email: orgEmail}
	if err := emailCfg.Get(); err != nil {
		beego.Info(err.Error())
		return nil, nil, err
	}

	ec, err := email.GetEmailClient(emailCfg.Platform)
	if err != nil {
		beego.Info(err.Error())
		return nil, nil, err
	}

	return emailCfg, ec, nil
}
