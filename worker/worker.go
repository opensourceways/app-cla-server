package worker

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"

	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

var worker IEmailWorker

type IEmailWorker interface {
	GenCLAPDFForCorporationAndSendIt(string, string, models.OrgInfo, models.CorporationSigning, []models.CLAField)
	SendSimpleMessage(orgEmail, platform, authorize string, msg *email.EmailMessage)
	Shutdown()
}

func GetEmailWorker() IEmailWorker {
	return worker
}

func InitEmailWorker(g pdf.IPDFGenerator) {
	worker = &emailWorker{
		pdfGenerator: g,
		stop:         make(chan struct{}),
	}
}

type emailWorker struct {
	pdfGenerator pdf.IPDFGenerator
	wg           sync.WaitGroup
	stop         chan struct{}
}

func (w *emailWorker) Shutdown() {
	close(w.stop)

	logs.Info("worker exit")

	// Handle remaining requests
	w.wg.Wait()
}

func (w *emailWorker) GenCLAPDFForCorporationAndSendIt(linkID, claFile string, orgInfo models.OrgInfo, signing models.CorporationSigning, claFields []models.CLAField) {
	f := func() {
		defer func() {
			w.wg.Done()
		}()

		emailCfg, ec, err := getEmailClient(orgInfo.OrgEmail)
		if err != nil {
			logs.Error("get email client failed, err:", err)
			return
		}

		data := email.CorporationSigning{
			Org:         orgInfo.OrgAlias,
			Date:        signing.Date,
			AdminName:   signing.AdminName,
			ProjectURL:  orgInfo.ProjectURL(),
			SigningInfo: buildCorpSigningInfo(&signing, claFields),
		}

		file := ""
		fileExist := func() bool {
			return file != "" && !util.IsFileNotExist(file)
		}

		defer func() {
			if fileExist() {
				os.Remove(file)
			}
		}()

		genFile := func() error {
			if fileExist() {
				return nil
			}

			file, err = w.pdfGenerator.GenPDFForCorporationSigning(linkID, claFile, &signing, claFields)
			if err != nil {
				return fmt.Errorf("error to generate pdf, err: %s", err.Error())
			}

			return nil
		}

		var msg *email.EmailMessage
		genMsg := func() error {
			if msg != nil {
				return nil
			}

			if msg, err = data.GenEmailMsg(); err != nil {
				return fmt.Errorf("error to gen email msg, err:%s", err.Error())
			}

			msg.Subject = fmt.Sprintf("Signing Corporation CLA on project of \"%s\"", data.Org)
			msg.To = []string{signing.AdminEmail}

			return nil
		}

		action := func() error {
			if err := genMsg(); err != nil {
				return err
			}

			if err := genFile(); err != nil {
				return err
			}

			msg.Attachment = file

			if err := ec.SendEmail(emailCfg.Token, emailCfg.Authorize, msg); err != nil {
				return fmt.Errorf("error to send email, err:%s", err.Error())
			}
			return nil
		}

		index := fmt.Sprintf(
			"sending email to %s/%s/%s:%s.",
			orgInfo.Platform, orgInfo.OrgID, orgInfo.RepoID,
			util.EmailSuffix(signing.AdminEmail),
		)

		w.tryToSendEmail(func() error {
			if err := action(); err != nil {
				return fmt.Errorf("%s %s", index, err.Error())
			}
			return nil
		})
	}

	w.wg.Add(1)
	go f()
}

func (w *emailWorker) SendSimpleMessage(orgEmail, platform, authorize string, msg *email.EmailMessage) {
	f := func() {
		defer func() {
			w.wg.Done()
		}()

		emailCfg, ec, err := getEmailClient(orgEmail)
		if err != nil {
			return
		}

		action := func() error {
			if err := ec.SendEmail(emailCfg.Token, emailCfg.Authorize, msg); err != nil {
				return fmt.Errorf("error to send email, err:%s", err.Error())
			}
			return nil
		}

		w.tryToSendEmail(action)
	}

	d := func() {
		defer func() {
			w.wg.Done()
		}()
		ec, err := email.EmailAgent.GetEmailClient(platform)
		if err != nil {
			return
		}

		action := func() error {
			if err := ec.SendEmail(nil, authorize, msg); err != nil {
				return fmt.Errorf("error to send email, err:%s", err.Error())
			}
			return nil
		}

		w.tryToSendEmail(action)
	}

	w.wg.Add(1)
	if authorize != "" {
		go d()
	} else {
		go f()
	}

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

func getEmailClient(orgEmail string) (*models.OrgEmail, email.IEmail, error) {
	emailCfg, merr := models.GetOrgEmailInfo(orgEmail)
	if merr != nil {
		logs.Info(merr.Error())
		return nil, nil, merr
	}

	ec, err := email.EmailAgent.GetEmailClient(emailCfg.Platform)
	if err != nil {
		logs.Info(err.Error())
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
