package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewCorpEmailDomainAdapter(s app.CorpEmailDomainService) *corpEmailDomainAdatper {
	return &corpEmailDomainAdatper{s}
}

type corpEmailDomainAdatper struct {
	s app.CorpEmailDomainService
}

func (adapter *corpEmailDomainAdatper) Verify(
	csId string, email string,
) (string, models.IModelError) {
	cmd, err := adapter.cmdToVerifyEmailDomain(csId, email)
	if err != nil {
		return "", toModelError(err)
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
		return toModelError(err)
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

	return
}

func (adapter *corpEmailDomainAdatper) List(csId string) ([]string, models.IModelError) {
	v, err := adapter.s.List(csId)
	if err != nil {
		return nil, toModelError(err)
	}

	return v, nil
}
