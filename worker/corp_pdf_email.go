package worker

import (
	"fmt"
	"os"
	"strings"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
	"github.com/opensourceways/app-cla-server/util"
)

func newCorpPDFEmail(
	linkID string,
	orgInfo *models.OrgInfo,
	claInfo *models.CLAInfo,
	signing *models.CorporationSigning,
) *corpPDFEmail {
	return &corpPDFEmail{
		linkID:  linkID,
		claInfo: *claInfo,
		orgInfo: *orgInfo,
		signing: *signing,
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

type corpPDFEmail struct {
	msg         *EmailMessage
	tmpl        emailtmpl.CorporationSigning
	tmplDone    bool
	pdfFilePath string

	linkID  string
	claInfo models.CLAInfo
	orgInfo models.OrgInfo
	signing models.CorporationSigning
}

func (impl *corpPDFEmail) do() error {
	if err := impl.genFile(); err != nil {
		return err
	}

	if err := impl.genMsg(); err != nil {
		return err
	}

	err := emailservice.SendEmail(
		impl.orgInfo.OrgEmailPlatform, impl.msg,
	)
	if err != nil {
		return fmt.Errorf("error to send email, err:%s", err.Error())
	}

	return nil
}

func (impl *corpPDFEmail) clean() {
	if impl.fileExist() {
		os.Remove(impl.pdfFilePath)
	}
}

func (impl *corpPDFEmail) fileExist() bool {
	return impl.pdfFilePath != "" && !util.IsFileNotExist(impl.pdfFilePath)
}

func (impl *corpPDFEmail) genFile() error {
	if impl.fileExist() {
		return nil
	}

	v, err := pdfGenerator.GenPDFForCorporationSigning(
		impl.linkID, impl.claInfo.CLAFile, &impl.signing, impl.claInfo.Fields,
	)
	if err != nil {
		return fmt.Errorf("error to generate pdf, err: %s", err.Error())
	}

	impl.pdfFilePath = v

	return nil
}

func (impl *corpPDFEmail) genEmailTmpl() {
	if impl.tmplDone {
		return
	}

	orgInfo := &impl.orgInfo
	signing := &impl.signing

	impl.tmpl = emailtmpl.CorporationSigning{
		Org:         orgInfo.OrgAlias,
		Date:        signing.Date,
		AdminName:   signing.AdminName,
		ProjectURL:  orgInfo.ProjectURL(),
		SigningInfo: buildCorpSigningInfo(signing, impl.claInfo.Fields),
	}

	impl.tmplDone = true
}

func (impl *corpPDFEmail) genMsg() error {
	if impl.msg != nil {
		return nil
	}

	impl.genEmailTmpl()

	msg, err := impl.tmpl.GenEmailMsg()
	if err != nil {
		return fmt.Errorf("error to gen email msg, err:%s", err.Error())
	}

	msg.Subject = fmt.Sprintf(
		"Signing Corporation CLA on project of \"%s\"", impl.orgInfo.OrgAlias,
	)
	msg.To = []string{impl.signing.AdminEmail}
	msg.From = impl.orgInfo.OrgEmail
	msg.Attachment = impl.pdfFilePath

	impl.msg = msg

	return nil
}
