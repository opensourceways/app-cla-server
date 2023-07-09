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
	vcService vcservice.VCService,
	es emailcredential.EmailCredential,
) SMTPService {
	return &smtpService{
		vc: vcService,
		es: es,
	}
}

// smtpService
type smtpService struct {
	vc vcservice.VCService
	es emailcredential.EmailCredential
}

func (s *smtpService) Verify(cmd *CmdToVerifySMTPEmail) (string, error) {
	p, err := cmd.purpose()
	if err != nil {
		return "", err
	}

	return s.vc.New(p)
}

func (s *smtpService) Authorize(cmd *CmdToAuthorizeSMTPEmail) error {
	k, err := cmd.key()
	if err != nil {
		return err
	}

	if err := s.vc.Verify(&k); err != nil {
		return err
	}

	v := cmd.emailCredential()
	return s.es.Add(&v)
}
