package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/smtpimpl"
)

func NewSMTPAdapter(
	s app.SMTPService,
) *smtpAdapter {
	return &smtpAdapter{s: s}
}

type smtpAdapter struct {
	s app.SMTPService
}

func (adapter *smtpAdapter) Verify(opt *models.EmailAuthorizationReq) (string, models.IModelError) {
	cmd, err := adapter.cmdToVerifySMTPEmail(opt)
	if err != nil {
		return "", toModelError(err)
	}

	v, err := adapter.s.Verify(&cmd)
	if err != nil {
		return "", toModelError(err)
	}

	return v, nil
}

func (adapter *smtpAdapter) Authorize(opt *models.EmailAuthorization) models.IModelError {
	cmd, err := adapter.cmdToAuthorizeSMTPEmail(opt)
	if err != nil {
		return toModelError(err)
	}

	if err := adapter.s.Authorize(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *smtpAdapter) cmdToAuthorizeSMTPEmail(opt *models.EmailAuthorization) (
	cmd app.CmdToAuthorizeSMTPEmail, err error,
) {
	v, err := adapter.cmdToVerifySMTPEmail(&opt.EmailAuthorizationReq)
	if err == nil {
		cmd.CmdToVerifySMTPEmail = v
		cmd.VerificationCode = opt.Code
	}

	return
}

func (adapter *smtpAdapter) cmdToVerifySMTPEmail(opt *models.EmailAuthorizationReq) (
	cmd app.CmdToVerifySMTPEmail, err error,
) {
	if cmd.EmailAddr, err = dp.NewEmailAddr(opt.Email); err != nil {
		return
	}

	cmd.Code = opt.Authorize
	cmd.Platform = smtpimpl.Platform()

	return
}
