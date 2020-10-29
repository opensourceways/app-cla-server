package worker

import (
	"fmt"
	"os"
	"strings"
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

		data := email.CorporationSigning{
			AdminName:   signing.AdminName,
			Org:         orgCLA.OrgID,
			Project:     util.ProjectName(orgCLA.OrgID, orgCLA.RepoID),
			Date:        signing.Date,
			SingingInfo: buildCorpSigningInfo(signing, cla),
		}

		var msg *email.EmailMessage
		file := ""

		for i := 0; i < 10; i++ {
			if this.shutdown {
				beego.Info("email worker exits forcedly")
				break
			}

			var err error

			if msg == nil {
				if msg, err = data.GenEmailMsg(); err != nil {
					next(err)
					continue
				}
				msg.Subject = fmt.Sprintf("Signing Corporation CLA on project of \"%s\"", data.Project)
				msg.To = []string{signing.AdminEmail}
			}

			if file == "" || util.IsFileNotExist(file) {
				file, err = this.pdfGenerator.GenPDFForCorporationSigning(orgCLA, signing, cla)
				if err != nil {
					next(fmt.Errorf(
						"Failed to generate pdf for corp signing(%s:%s:%s/%s): %s",
						orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID, util.EmailSuffix(signing.AdminEmail),
						err.Error()))
					continue
				}
			}
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

		for i := 0; i < 10; i++ {
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

func buildCorpSigningInfo(signing *models.CorporationSigning, cla *models.CLA) string {
	orders, keys, err := pdf.BuildCorpContact(cla)
	if err != nil {
		return ""
	}

	v := make([]string, 0, len(orders))
	for _, i := range orders {
		v = append(v, fmt.Sprintf("%s: %s", keys[i], signing.Info[i]))
	}

	return strings.Join(v, "\n")
}
