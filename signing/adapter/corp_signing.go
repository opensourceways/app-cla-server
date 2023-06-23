package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewCorpSigningAdapter(s app.CorpSigningService) *corpSigningAdatper {
	return &corpSigningAdatper{s}
}

type corpSigningAdatper struct {
	s app.CorpSigningService
}

func (adapter *corpSigningAdatper) Sign(opt *models.CorporationSigningCreateOption, linkId string) models.IModelError {
	cmd, err := adapter.cmdToSignCorpCLA(opt, linkId)
	if err != nil {
		return toModelError(err)
	}

	if err = adapter.s.Sign(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *corpSigningAdatper) cmdToSignCorpCLA(opt *models.CorporationSigningCreateOption, linkId string) (
	cmd app.CmdToSignCorpCLA, err error,
) {
	cmd.Link.Id = linkId
	// missing cla id
	if cmd.Link.Language, err = dp.NewLanguage(opt.CLALanguage); err != nil {
		return
	}

	if cmd.CorpName, err = dp.NewCorpName(opt.CorporationName); err != nil {
		return
	}

	if cmd.Representative.Name, err = dp.NewName(opt.AdminName); err != nil {
		return
	}

	if cmd.Representative.EmailAddr, err = dp.NewEmailAddr(opt.AdminEmail); err != nil {
		return
	}

	cmd.AllSingingInfo = opt.Info

	return
}
