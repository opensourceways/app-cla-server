package worker

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
	"github.com/opensourceways/app-cla-server/util"
)

var worker IEmailWorker

type EmailMessage = emailservice.EmailMessage

type IEmailWorker interface {
	GenCLAPDFForCorporationAndSendIt(string, string, models.OrgInfo, models.CorporationSigning, []models.CLAField)
	SendSimpleMessage(platform string, msg *EmailMessage)
	Shutdown()
}

func GetEmailWorker() IEmailWorker {
	return worker
}

func Init(g pdf.IPDFGenerator) {
	worker = &emailWorker{
		pdfGenerator: g,
		stop:         make(chan struct{}),
	}
}

func Exit() {
	if worker != nil {
		worker.Shutdown()
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

func (w *emailWorker) GenCLAPDFForCorporationAndSendIt(
	linkID, claFile string, orgInfo models.OrgInfo,
	signing models.CorporationSigning,
	claFields []models.CLAField,
) {
	f := func() {
		defer func() {
			w.wg.Done()
		}()

		data := emailtmpl.CorporationSigning{
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

		var err error

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

		var msg *EmailMessage
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
			msg.From = orgInfo.OrgEmail
			if err := emailservice.SendEmail(orgInfo.OrgEmailPlatform, msg); err != nil {
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

func (w *emailWorker) SendSimpleMessage(emailPlatform string, msg *EmailMessage) {
	f := func() {
		defer func() {
			w.wg.Done()
		}()

		action := func() error {
			if err := emailservice.SendEmail(emailPlatform, msg); err != nil {
				return fmt.Errorf("error to send email, err:%s", err.Error())
			}
			return nil
		}

		w.tryToSendEmail(action)
	}

	w.wg.Add(1)
	go f()
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

func buildCorpSigningInfo(signing *models.CorporationSigning, claFields []models.CLAField) string {
	orders, titles := pdf.BuildCorpContact(claFields)

	v := make([]string, 0, len(orders))
	for _, i := range orders {
		v = append(v, fmt.Sprintf("%s: %s", titles[i], signing.Info[i]))
	}
	v = append(v, fmt.Sprintf("Date: %s", signing.Date))

	return "  " + strings.Join(v, "\n  ")
}
