package adapter

import (
	"errors"
	"strings"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewCorpEmailDomainAdapter(
	s app.CorpEmailDomainService,
	invalidCorpEmailDomain []string,
) *corpEmailDomainAdatper {
	v := make([]string, len(invalidCorpEmailDomain))
	for i, item := range invalidCorpEmailDomain {
		v[i] = strings.ToLower(item)
	}

	return &corpEmailDomainAdatper{
		s:                      s,
		invalidCorpEmailDomain: v,
	}
}

type corpEmailDomainAdatper struct {
	s                      app.CorpEmailDomainService
	invalidCorpEmailDomain []string
}

func (adapter *corpEmailDomainAdatper) isValidaCorpEmailDomain(v string) bool {
	v = strings.ToLower(v)

	for _, item := range adapter.invalidCorpEmailDomain {
		if item == v {
			return false
		}
	}

	return true
}

func (adapter *corpEmailDomainAdatper) Verify(
	csId string, email string,
) (string, models.IModelError) {
	cmd, err := adapter.cmdToVerifyEmailDomain(csId, email)
	if err != nil {
		return "", errBadRequestParameter(err)
	}

	v, err := adapter.s.Verify(&cmd)
	if err != nil {
		return "", toModelError(err)
	}

	return v, nil
}

func (adapter *corpEmailDomainAdatper) Add(
	csId string, opt *models.CorpEmailDomainCreateOption,
) models.IModelError {
	v, err := adapter.cmdToVerifyEmailDomain(csId, opt.SubEmail)
	if err != nil {
		return errBadRequestParameter(err)
	}

	cmd := app.CmdToAddEmailDomain{CmdToVerifyEmailDomain: v}
	cmd.VerificationCode = opt.VerificationCode

	if err = adapter.s.Add(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *corpEmailDomainAdatper) cmdToVerifyEmailDomain(csId string, email string) (
	cmd app.CmdToVerifyEmailDomain, err error,
) {
	cmd = app.CmdToVerifyEmailDomain{
		CorpSigningId: csId,
	}

	cmd.EmailAddr, err = dp.NewEmailAddr(email)
	if err != nil {
		return
	}

	if !adapter.isValidaCorpEmailDomain(cmd.EmailAddr.Domain()) {
		err = errors.New("invalid email domain")
		return
	}

	return
}

func (adapter *corpEmailDomainAdatper) List(csId string) ([]string, models.IModelError) {
	v, err := adapter.s.List(csId)
	if err != nil {
		return nil, toModelError(err)
	}

	return v, nil
}
