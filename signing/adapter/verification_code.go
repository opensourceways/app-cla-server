package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewVerificationCodeAdapter(s app.VerificationCodeService) *verificationCodeAdatper {
	return &verificationCodeAdatper{s}
}

type verificationCodeAdatper struct {
	s app.VerificationCodeService
}

func (adapter *verificationCodeAdatper) CreateForSigning(linkId string, email string) (
	string, models.IModelError,
) {
	cmd := app.CmdToCreateCodeForSigning{
		LinkId: linkId,
	}

	var err error

	if cmd.EmailAddr, err = dp.NewEmailAddr(email); err != nil {
		return "", toModelError(err)
	}

	code, err := adapter.s.CreateForSigning(&cmd)
	if err != nil {
		return "", toModelError(err)
	}

	return code, nil
}

func (adapter *verificationCodeAdatper) ValidateForSigning(linkId string, email, code string) models.IModelError {
	cmd := app.CmdToValidateCodeForSigning{
		Code: code,
	}
	cmd.LinkId = linkId

	var err error

	if cmd.EmailAddr, err = dp.NewEmailAddr(email); err != nil {
		return toModelError(err)
	}

	if err := adapter.s.ValidateForSigning(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *verificationCodeAdatper) CreateForAddingEmailDomain(csId string, email string) (
	string, models.IModelError,
) {
	cmd := app.CmdToCreateCodeForEmailDomain{
		CorpSigningId: csId,
	}

	var err error

	if cmd.EmailAddr, err = dp.NewEmailAddr(email); err != nil {
		return "", toModelError(err)
	}

	code, err := adapter.s.CreateForAddingEmailDomain(&cmd)
	if err != nil {
		return "", toModelError(err)
	}

	return code, nil
}

func (adapter *verificationCodeAdatper) ValidateForAddingEmailDomain(
	csId string, email, code string,
) models.IModelError {

	cmd := app.CmdToValidateCodeForEmailDomain{
		Code: code,
	}
	cmd.CorpSigningId = csId

	var err error

	if cmd.EmailAddr, err = dp.NewEmailAddr(email); err != nil {
		return toModelError(err)
	}

	if err := adapter.s.ValidateForAddingEmailDomain(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}
