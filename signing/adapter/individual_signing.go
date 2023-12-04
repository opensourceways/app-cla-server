package adapter

import (
	"errors"
	"strings"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewIndividualSigningAdapter(
	s app.IndividualSigningService,
	allowedEmailDomains []string,
) *individualSigningAdatper {
	return &individualSigningAdatper{
		s:              s,
		emailValidator: newEmailValidator(allowedEmailDomains),
	}
}

type individualSigningAdatper struct {
	s              app.IndividualSigningService
	emailValidator emailValidator
}

func (adapter *individualSigningAdatper) checkEmail(email string) (dp.EmailAddr, models.IModelError) {
	return adapter.emailValidator.validate(email, true)
}

func (adapter *individualSigningAdatper) Verify(linkId, email string) (string, models.IModelError) {
	v, err := adapter.checkEmail(email)
	if err != nil {
		return "", err
	}

	return createCodeForSigning(linkId, v, adapter.s.Verify)
}

// Sign
func (adapter *individualSigningAdatper) Sign(
	linkId string, opt *models.IndividualSigning, claFields []models.CLAField,
) models.IModelError {
	cmd, err := adapter.cmdToSignIndividualCLA(linkId, opt, claFields)
	if err != nil {
		return errBadRequestParameter(err)
	}

	if err = adapter.s.Sign(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *individualSigningAdatper) cmdToSignIndividualCLA(
	linkId string, opt *models.IndividualSigning, claFields []models.CLAField,
) (
	cmd app.CmdToSignIndividualCLA, err error,
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

	if cmd.Rep.Name, err = dp.NewName(opt.Name); err != nil {
		return
	}

	if cmd.Rep.EmailAddr, err = adapter.checkEmail(opt.Email); err != nil {
		return
	}

	cmd.AllSingingInfo, err = getAllSigningInfo(
		opt.Info, claFields, dp.CLATypeIndividual, cmd.Link.Language,
	)
	if err != nil {
		return
	}

	cmd.VerificationCode = opt.VerificationCode

	return
}

// Check
func (adapter *individualSigningAdatper) Check(linkId string, email string) (bool, models.IModelError) {
	cmd := app.CmdToCheckSinging{
		LinkId: linkId,
	}

	e, me := adapter.checkEmail(email)
	if me != nil {
		return false, me
	}
	cmd.EmailAddr = e

	v, err := adapter.s.Check(&cmd)
	if err != nil {
		return v, toModelError(err)
	}

	return v, nil
}

// emailValidator
type emailValidator map[string]bool

func newEmailValidator(v []string) emailValidator {
	m := map[string]bool{}
	for _, v := range v {
		m[v] = true
	}

	return emailValidator(m)
}

func (ev emailValidator) validate(email string, expect bool) (dp.EmailAddr, models.IModelError) {
	v, err := dp.NewEmailAddr(email)
	if err != nil {
		return nil, models.NewModelError(models.ErrNotAnEmail, err)
	}

	b := ev[strings.ToLower(v.Domain())]

	if (expect && !b) || (!expect && b) {
		return nil, models.NewModelError(
			models.ErrRestrictedEmailSuffix,
			errors.New("not allowed email domain"),
		)
	}

	return v, nil
}

func createCodeForSigning(
	index string, email dp.EmailAddr,
	f func(*app.CmdToCreateVerificationCode) (string, error),
) (
	string, models.IModelError,
) {
	code, err := f(&app.CmdToCreateVerificationCode{
		Id:        index,
		EmailAddr: email,
	})
	if err != nil {
		return "", toModelError(err)
	}

	return code, nil
}
