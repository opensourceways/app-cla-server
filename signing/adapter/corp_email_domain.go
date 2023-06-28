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

func (adapter *corpEmailDomainAdatper) Add(
	csId string, opt *models.CorpEmailDomainCreateOption,
) models.IModelError {
	cmd := app.CmdToAddEmailDomain{
		CorpSigningId: csId,
	}

	var err error

	if cmd.EmailAddr, err = dp.NewEmailAddr(opt.SubEmail); err != nil {
		return toModelError(err)
	}

	if err = adapter.s.Add(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *corpEmailDomainAdatper) List(csId string) ([]string, models.IModelError) {
	v, err := adapter.s.List(csId)
	if err != nil {
		return nil, toModelError(err)
	}

	return v, nil
}
