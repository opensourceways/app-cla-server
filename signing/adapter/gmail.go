package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
)

func NewGmailAdapter(s app.GmailService) *gmailAdatper {
	return &gmailAdatper{s: s}
}

type gmailAdatper struct {
	s app.GmailService
}

func (adapter *gmailAdatper) Authorize(code, scope string) (string, models.IModelError) {
	v, err := adapter.s.Authorize(&app.CmdToAuthorizeGmail{
		Code:  code,
		Scope: scope,
	})

	if err != nil {
		return "", toModelError(err)
	}

	return v, nil
}
