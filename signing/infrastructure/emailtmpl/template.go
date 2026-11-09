package emailtmpl

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/opensourceways/obs-sdk/go/log"

	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
	"github.com/opensourceways/app-cla-server/util"
)

const (
	TmplCorporationSigning    = "corporation signing"
	TmplIndividualSigning     = "individual signing"
	TmplEmployeeSigning       = "employee signing"
	TmplNotifyingManager      = "notifying manager"
	TmplVerificationCode      = "verification code"
	TmplAddingCorpEmailDomain = "adding corp email domain"
	TmplAddingCorpAdmin       = "adding corp admin"
	TmplAddingCorpManager     = "adding corp manager"
	TmplRemovingCorpManager   = "removing corp manager"
	TmplActivatingEmployee    = "activating employee"
	TmplInactivaingEmployee   = "inactivating employee"
	TmplRemovingingEmployee   = "removing employee"
	TmplPasswordRetrieval     = "password retrieval"
	TmplEmailVerification     = "email verification"
	TmplCLAUpdated            = "cla updated"
	TmplCLAUpdatedIndividual  = "cla updated individual"
	TmplAutoApprovalEnabled   = "auto approval enabled"
	TmplAutoApprovalDisabled  = "auto approval disabled"
	TmplAutoApprovalEmployee  = "auto approval employee"
	TmplAutoApprovalManager   = "auto approval manager"
)

const (
	langEN = "en"
	langZH = "zh"

	tmplDir = "./conf/email-template/"
)

var msgTmpl = map[string]map[string]*template.Template{}

type EmailMessage = emailservice.EmailMessage

var tmplFiles = map[string]string{
	TmplCorporationSigning:    "corporation-signing.tmpl",
	TmplIndividualSigning:     "individual-signing.tmpl",
	TmplEmployeeSigning:       "employee-signing.tmpl",
	TmplNotifyingManager:      "notifying-corp-manager.tmpl",
	TmplVerificationCode:      "verification-code.tmpl",
	TmplAddingCorpEmailDomain: "adding-corp-email-domain.tmpl",
	TmplAddingCorpAdmin:       "adding-corp-admin.tmpl",
	TmplAddingCorpManager:     "adding-corp-manager.tmpl",
	TmplRemovingCorpManager:   "removing-corp-manager.tmpl",
	TmplActivatingEmployee:    "activating-employee.tmpl",
	TmplInactivaingEmployee:   "inactivating-employee.tmpl",
	TmplRemovingingEmployee:   "removing-employee.tmpl",
	TmplPasswordRetrieval:     "password-retrieval.tmpl",
	TmplEmailVerification:     "email_verification.tmpl",
	TmplCLAUpdated:            "cla-updated.tmpl",
	TmplCLAUpdatedIndividual:  "cla-updated-individual.tmpl",
	TmplAutoApprovalEnabled:   "auto-approval-enabled.tmpl",
	TmplAutoApprovalDisabled:  "auto-approval-disabled.tmpl",
	TmplAutoApprovalEmployee:  "auto-approval-employee.tmpl",
	TmplAutoApprovalManager:   "auto-approval-manager.tmpl",
}

func Init() error {
	for name, file := range tmplFiles {
		msgTmpl[name] = map[string]*template.Template{}

		enPath := tmplDir + langEN + "/" + file
		enTmpl, err := util.NewTemplate(name, enPath)
		if err != nil {
			return err
		}
		msgTmpl[name][langEN] = enTmpl

		zhPath := tmplDir + langZH + "/" + file
		if _, err := os.Stat(zhPath); err == nil {
			zhTmpl, err := util.NewTemplate(name, zhPath)
			if err != nil {
				return err
			}
			msgTmpl[name][langZH] = zhTmpl
		} else {
			log.Warn("zh email template not found, fallback to en", "template", name, "path", zhPath)
		}
	}

	return nil
}

func NormalizeLang(v string) string {
	switch v {
	case "Chinese", langZH:
		return langZH
	case "English", langEN:
		return langEN
	default:
		return langEN
	}
}

func findTmpl(name, lang string) *template.Template {
	if m, ok := msgTmpl[name]; ok {
		if t, ok2 := m[lang]; ok2 {
			return t
		}
		return m[langEN]
	}
	return nil
}

func genEmailMsg(tmplName, lang string, data interface{}) (msg EmailMessage, err error) {
	tmpl := findTmpl(tmplName, NormalizeLang(lang))
	if tmpl == nil {
		err = fmt.Errorf("failed to generate email msg: didn't find msg template: %s", tmplName)

		return
	}

	var subject bytes.Buffer
	if err = tmpl.ExecuteTemplate(&subject, "subject", data); err != nil {
		err = fmt.Errorf("failed to render email subject: %s", err.Error())

		return
	}
	msg.Subject = subject.String()

	var body bytes.Buffer
	if err = tmpl.ExecuteTemplate(&body, "body", data); err != nil {
		err = fmt.Errorf("failed to render email body: %s", err.Error())

		return
	}
	msg.Content = body

	return
}

type CorporationSigning struct {
	Lang        string
	Org         string
	Date        string
	AdminName   string
	ProjectURL  string
	SigningInfo string
}

func (data *CorporationSigning) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplCorporationSigning, data.Lang, data)
}

type IndividualSigning struct {
	Lang string
	Name string
}

func (data *IndividualSigning) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplIndividualSigning, data.Lang, data)
}

type VerificationCode struct {
	Lang        string
	Email       string
	Org         string
	Code        string
	ProjectURL  string
}

func (data *VerificationCode) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplVerificationCode, data.Lang, data)
}

type AddingCorpEmailDomain struct {
	Lang        string
	Corp        string
	Org         string
	Code        string
	ProjectURL  string
}

func (cse AddingCorpEmailDomain) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplAddingCorpEmailDomain, cse.Lang, cse)
}

type AddingCorpManager struct {
	Lang             string
	Admin            bool
	ID               string
	User             string
	Email            string
	Password         []byte
	Org              string
	ProjectURL       string
	URLOfCLAPlatform string
}

func (data *AddingCorpManager) GenEmailMsg() (msg EmailMessage, err error) {
	if data.Admin {
		msg, err = genEmailMsg(TmplAddingCorpAdmin, data.Lang, data)
	} else {
		msg, err = genEmailMsg(TmplAddingCorpManager, data.Lang, data)
	}

	msg.HasSecret = true

	return
}

type RemovingCorpManager struct {
	Lang        string
	User        string
	Org         string
	ProjectURL  string
}

func (data *RemovingCorpManager) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplRemovingCorpManager, data.Lang, data)
}

type EmployeeSigning struct {
	Lang        string
	Name        string
	Org         string
	ProjectURL  string
	Managers    string
}

func (data *EmployeeSigning) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplEmployeeSigning, data.Lang, data)
}

type NotifyingManager struct {
	Lang             string
	EmployeeEmail    string
	ProjectURL       string
	URLOfCLAPlatform string
	Org              string
}

func (data *NotifyingManager) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplNotifyingManager, data.Lang, data)
}

type EmployeeNotification struct {
	Lang     string
	Removing bool
	Active   bool
	Inactive bool

	Name       string
	ProjectURL string
	Manager    string
	Org        string
}

func (data *EmployeeNotification) GenEmailMsg() (EmailMessage, error) {
	if data.Active {
		return genEmailMsg(TmplActivatingEmployee, data.Lang, data)
	}

	if data.Inactive {
		return genEmailMsg(TmplInactivaingEmployee, data.Lang, data)
	}

	if data.Removing {
		return genEmailMsg(TmplRemovingingEmployee, data.Lang, data)
	}

	return EmailMessage{}, fmt.Errorf("do nothing")
}

type PasswordRetrieval struct {
	Lang         string
	Timeout      int64
	Org          string
	ResetURL     string
	RetrievalURL string
}

func (p PasswordRetrieval) GenEmailMsg() (msg EmailMessage, err error) {
	msg, err = genEmailMsg(TmplPasswordRetrieval, p.Lang, p)
	if err != nil {
		return msg, err
	}

	msg.MIME = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	return msg, nil
}

type EmailVerification struct {
	Lang string
	Code string
}

func (data *EmailVerification) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplEmailVerification, data.Lang, data)
}

type CLAUpdated struct {
	Lang             string
	Org              string
	AdminName        string
	ProjectURL       string
	URLOfCLAPlatform string
}

func (data *CLAUpdated) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplCLAUpdated, data.Lang, data)
}

type CLAUpdatedIndividual struct {
	Lang             string
	Name             string
	Org              string
	UpdateDate       string
	GracePeriodDays  int
	SignCLAURL       string
	ProjectURL       string
}

func (data *CLAUpdatedIndividual) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplCLAUpdatedIndividual, data.Lang, data)
}

type AutoApprovalEnabled struct {
	Lang             string
	Org              string
	ProjectURL       string
	URLOfCLAPlatform string
}

func (data *AutoApprovalEnabled) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplAutoApprovalEnabled, data.Lang, data)
}

type AutoApprovalDisabled struct {
	Lang             string
	Org              string
	ProjectURL       string
	URLOfCLAPlatform string
}

func (data *AutoApprovalDisabled) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplAutoApprovalDisabled, data.Lang, data)
}

type AutoApprovalEmployee struct {
	Lang        string
	Name        string
	Org         string
	ProjectURL  string
}

func (data *AutoApprovalEmployee) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplAutoApprovalEmployee, data.Lang, data)
}

type AutoApprovalManager struct {
	Lang             string
	EmployeeEmail    string
	Org              string
	ProjectURL       string
	URLOfCLAPlatform string
}

func (data *AutoApprovalManager) GenEmailMsg() (EmailMessage, error) {
	return genEmailMsg(TmplAutoApprovalManager, data.Lang, data)
}
