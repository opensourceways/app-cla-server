package email

import (
	"fmt"
	"text/template"

	"github.com/opensourceways/app-cla-server/util"
)

const (
	TmplCorporationSigning = "corporation signing"
	TmplIndividualSigning  = "individual signing"
)

var msgTmpl = map[string]*template.Template{}

func initTemplate() error {
	items := map[string]string{
		TmplCorporationSigning: "./conf/email-template/corporation-signing.tmpl",
		TmplIndividualSigning:  "./conf/email-template/individual-signing.tmpl",
	}

	for name, path := range items {
		tmpl, err := util.NewTemplate(name, path)
		if err != nil {
			return err
		}
		msgTmpl[name] = tmpl
	}

	return nil
}

func findTmpl(name string) *template.Template {
	v, ok := msgTmpl[name]
	if ok {
		return v
	}
	return nil
}

func genEmailMsg(tmplName string, data interface{}) (*EmailMessage, error) {
	tmpl := findTmpl(tmplName)
	if tmpl == nil {
		return nil, fmt.Errorf("Failed to generate email msg: didn't find msg template: %s", tmplName)
	}

	str, err := util.RenderTemplate(tmpl, data)
	if err != nil {
		return nil, err
	}
	return &EmailMessage{Content: str}, nil
}

type IEmailMessageBulder interface {
	// msg returned only includes content and subject
	GenEmailMsg() (*EmailMessage, error)
}

type CorporationSigning struct{}

func (this CorporationSigning) GenEmailMsg() (*EmailMessage, error) {
	return genEmailMsg(TmplCorporationSigning, this)
}

type IndividualSigning struct{}

func (this IndividualSigning) GenEmailMsg() (*EmailMessage, error) {
	return genEmailMsg(TmplIndividualSigning, this)
}
