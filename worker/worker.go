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
	GenCLAPDFForCorporationAndSendIt(string, string, models.OrgInfo, models.CorporationSigning, []models.CLAField)
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

func (this *emailWorker) GenCLAPDFForCorporationAndSendIt(linkID, claFile string, orgInfo models.OrgInfo, signing models.CorporationSigning, claFields []models.CLAField) {
	f := func() {
		defer func() {
			this.wg.Done()
		}()

		emailCfg, ec, err := getEmailClient(orgInfo.OrgEmail)
		if err != nil {
			return
		}

		data := email.CorporationSigning{
			Org:         orgInfo.OrgAlias,
			Date:        signing.Date,
			AdminName:   signing.AdminName,
			ProjectURL:  orgInfo.ProjectURL(),
			SigningInfo: buildCorpSigningInfo(&signing, claFields),
		}

		var msg *email.EmailMessage
		file := ""

		defer func() {
			if !util.IsFileNotExist(file) {
				os.Remove(file)
			}
		}()

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
				msg.Subject = fmt.Sprintf("Signing Corporation CLA on project of \"%s\"", data.Org)
				msg.To = []string{signing.AdminEmail}
			}

			if file == "" || util.IsFileNotExist(file) {
				file, err = this.pdfGenerator.GenPDFForCorporationSigning(linkID, claFile, &signing, claFields)
				if err != nil {
					next(fmt.Errorf(
						"Failed to generate pdf for corp signing(%s:%s:%s/%s): %s",
						orgInfo.Platform, orgInfo.OrgID, orgInfo.RepoID, util.EmailSuffix(signing.AdminEmail),
						err.Error()))
					continue
				}
			}
			msg.Attachment = file

			if err := ec.SendEmail(emailCfg.Token, msg); err != nil {
				next(err)
			} else {
				break
			}
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
	emailCfg, merr := models.GetOrgEmailInfo(orgEmail)
	if merr != nil {
		beego.Info(merr.Error())
		return nil, nil, merr
	}

	ec, err := email.EmailAgent.GetEmailClient(emailCfg.Platform)
	if err != nil {
		beego.Info(err.Error())
		return nil, nil, err
	}

	return emailCfg, ec, nil
}

func buildCorpSigningInfo(signing *models.CorporationSigning, claFields []models.CLAField) string {
	orders, titles := pdf.BuildCorpContact(claFields)

	v := make([]string, 0, len(orders))
	for _, i := range orders {
		v = append(v, fmt.Sprintf("%s: %s", titles[i], signing.Info[i]))
	}
	v = append(v, fmt.Sprintf("Date: %s", signing.Date))

	return "  " + strings.Join(v, "\n  ")
}
