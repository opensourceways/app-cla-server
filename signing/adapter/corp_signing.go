package adapter

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
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
	// TODO missing cla id
	cmd.Link.CLAId = opt.CLALanguage
	if cmd.Link.Language, err = dp.NewLanguage(opt.CLALanguage); err != nil {
		return
	}

	if cmd.CorpName, err = dp.NewCorpName(opt.CorporationName); err != nil {
		return
	}

	if cmd.Rep.Name, err = dp.NewName(opt.AdminName); err != nil {
		return
	}

	if cmd.Rep.EmailAddr, err = dp.NewEmailAddr(opt.AdminEmail); err != nil {
		return
	}

	cmd.AllSingingInfo = opt.Info

	return
}

// Remove
func (adapter *corpSigningAdatper) Remove(csId string) models.IModelError {
	if err := adapter.s.Remove(csId); err != nil {
		return toModelError(err)
	}

	return nil
}

// Get
func (adapter *corpSigningAdatper) Get(csId string) (
	models.CorporationSigning, models.IModelError,
) {
	item, err := adapter.s.Get(csId)
	if err != nil {
		return models.CorporationSigning{}, toModelError(err)
	}

	return models.CorporationSigning{
		CorporationSigningBasicInfo: dbmodels.CorporationSigningBasicInfo{
			Date:            item.Date,
			AdminName:       item.RepName,
			AdminEmail:      item.RepEmail,
			CLALanguage:     item.Language,
			CorporationName: item.CorpName,
		},
		Info: item.AllInfo,
	}, nil
}

// List
func (adapter *corpSigningAdatper) List(linkId string) (
	[]models.CorporationSigningSummary, models.IModelError,
) {
	v, err := adapter.s.List(linkId)
	if err != nil {
		return nil, toModelError(err)
	}

	r := make([]models.CorporationSigningSummary, len(v))
	for i := range v {
		item := &v[i]

		r[i] = models.CorporationSigningSummary{
			CorporationSigningBasicInfo: dbmodels.CorporationSigningBasicInfo{
				Date:            item.Date,
				AdminName:       item.RepName,
				AdminEmail:      item.RepEmail,
				CLALanguage:     item.Language,
				CorporationName: item.CorpName,
			},
			Id:          item.Id,
			AdminAdded:  item.HasAdminAdded,
			PDFUploaded: item.HasPDFUploaded,
		}
	}

	return r, nil
}

// FindCorpSigningId
func (adapter *corpSigningAdatper) FindCorpSummary(linkId string, email string) (
	interface{}, models.IModelError,
) {
	cmd := app.CmdToFindCorpSummary{
		LinkId: linkId,
	}

	var err error
	if cmd.EmailAddr, err = dp.NewEmailAddr(email); err != nil {
		return false, toModelError(err)
	}

	v, err := adapter.s.FindCorpSummary(&cmd)
	if err != nil {
		return nil, toModelError(err)
	}

	return v, nil
}
