package adapter

import (
	"errors"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewCorpSigningAdapter(
	s app.CorpSigningService,
	invalidCorpEmailDomain []string,
) *corpSigningAdatper {
	return &corpSigningAdatper{
		s:              s,
		emailValidator: newEmailValidator(invalidCorpEmailDomain),
	}
}

type corpSigningAdatper struct {
	s app.CorpSigningService

	emailValidator
}

func (adapter *corpSigningAdatper) checkEmail(email string) (dp.EmailAddr, models.IModelError) {
	return adapter.emailValidator.validate(email, false)
}

func (adapter *corpSigningAdatper) Verify(linkId, email string) (string, models.IModelError) {
	v, err := adapter.checkEmail(email)
	if err != nil {
		return "", err
	}

	return createCodeForSigning(linkId, v, adapter.s.Verify)
}

func (adapter *corpSigningAdatper) Sign(
	linkId string, opt *models.CorporationSigningCreateOption, claFields []models.CLAField,
) models.IModelError {
	cmd, err := adapter.cmdToSignCorpCLA(linkId, opt, claFields)
	if err != nil {
		return errBadRequestParameter(err)
	}

	if err = adapter.s.Sign(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *corpSigningAdatper) cmdToSignCorpCLA(
	linkId string, opt *models.CorporationSigningCreateOption, claFields []models.CLAField,
) (
	cmd app.CmdToSignCorpCLA, err error,
) {
	if !opt.PrivacyChecked {
		err = errors.New("must agree to the privacy statement")

		return
	}

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

	if cmd.Rep.EmailAddr, err = adapter.checkEmail(opt.AdminEmail); err != nil {
		return
	}

	cmd.AllSingingInfo, err = getAllSigningInfo(
		opt.Info, claFields, dp.CLATypeCorp, cmd.Link.Language,
	)
	if err != nil {
		return
	}

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

func getAllSigningInfo(
	input models.TypeSigningInfo, fields []models.CLAField, t dp.CLAType, l dp.Language,
) (domain.AllSingingInfo, error) {
	m := map[string]*dp.CLAField{}
	whitelist := dp.GetCLAFileds(t, l)
	for i := range whitelist {
		item := &whitelist[i]
		m[item.Type] = item
	}

	r := domain.AllSingingInfo{}
	for i := range fields {
		field := &fields[i]

		if v, ok := input[field.ID]; !ok {
			if field.Required {
				return nil, errors.New("missing field value")
			}
		} else {
			if !m[field.Type].IsValidValue(v) {
				return nil, errors.New("invalid field value")
			}

			r[field.ID] = v
		}
	}

	return r, nil
}
