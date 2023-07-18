package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain/emailcredential"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

type SMTPService interface {
	Verify(cmd *CmdToVerifySMTPEmail) (string, error)
	Authorize(cmd *CmdToAuthorizeSMTPEmail) error
}

func NewSMTPService(
	vc vcservice.VCService,
	es emailcredential.EmailCredential,
) SMTPService {
	return &smtpService{
		vc: verificationCodeService{vc},
		es: es,
	}
}

// smtpService
type smtpService struct {
	vc verificationCodeService
	es emailcredential.EmailCredential
}

func (s *smtpService) Verify(cmd *CmdToVerifySMTPEmail) (string, error) {
	// can't invoke cmd.clear(), because it needs send email later
	return s.vc.newCode(cmd)
}

func (s *smtpService) Authorize(cmd *CmdToAuthorizeSMTPEmail) error {
	defer cmd.clear()

	if err := s.vc.validate(cmd, cmd.VerificationCode); err != nil {
		return err
	}

	v := cmd.emailCredential()

	return s.es.Add(&v)
}
