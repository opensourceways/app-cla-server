package adapter

import (
	"errors"
	"strings"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewCorpSigningAdapter(
	s app.CorpSigningService,
	invalidCorpEmailDomain []string,
) *corpSigningAdatper {
	v := make([]string, len(invalidCorpEmailDomain))
	for i, item := range invalidCorpEmailDomain {
		v[i] = strings.ToLower(item)
	}

	return &corpSigningAdatper{
		s:                      s,
		invalidCorpEmailDomain: v,
	}
}

type corpSigningAdatper struct {
	s                      app.CorpSigningService
	invalidCorpEmailDomain []string
}

func (adapter *corpSigningAdatper) isValidaCorpEmailDomain(v string) bool {
	v = strings.ToLower(v)

	for _, item := range adapter.invalidCorpEmailDomain {
		if item == v {
			return false
		}
	}

	return true
}

func (adapter *corpSigningAdatper) Verify(linkId, email string) (string, models.IModelError) {
	return createCodeForSigning(linkId, email, adapter.s.Verify)
}

func (adapter *corpSigningAdatper) Sign(
	linkId string, opt *models.CorporationSigningCreateOption,
) models.IModelError {
	cmd, err := adapter.cmdToSignCorpCLA(linkId, opt)
	if err != nil {
		return errBadRequestParameter(err)
	}

	if err = adapter.s.Sign(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *corpSigningAdatper) cmdToSignCorpCLA(
	linkId string, opt *models.CorporationSigningCreateOption,
) (
	cmd app.CmdToSignCorpCLA, err error,
) {
	cmd.Link.Id = linkId
	cmd.Link.CLAId = opt.CLAId
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

	if !adapter.isValidaCorpEmailDomain(cmd.Rep.EmailAddr.Domain()) {
		err = errors.New("invalid email domain")

		return
	}

	cmd.AllSingingInfo = opt.Info
	cmd.VerificationCode = opt.VerificationCode

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
		CorporationSigningBasicInfo: models.CorporationSigningBasicInfo{
			Date:            item.Date,
			AdminName:       item.RepName,
			AdminEmail:      item.RepEmail,
			CLAId:           item.CLAId,
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
			CorporationSigningBasicInfo: models.CorporationSigningBasicInfo{
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
		return false, errBadRequestParameter(err)
	}

	v, err := adapter.s.FindCorpSummary(&cmd)
	if err != nil {
		return nil, toModelError(err)
	}

	return v, nil
}
