package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

type VerificationCodeService interface {
	CreateForSigning(cmd *CmdToCreateCodeForSigning) (string, error)
	ValidateForSigning(cmd *CmdToValidateCodeForSigning) error

	CreateForAddingEmailDomain(cmd *CmdToCreateCodeForEmailDomain) (string, error)
	ValidateForAddingEmailDomain(cmd *CmdToValidateCodeForEmailDomain) error
}

func NewVerificationCodeService(
	vcService vcservice.VCService,
) VerificationCodeService {
	return &verificationCodeService{
		vcService: vcService,
	}
}

// verificationCodeService
type verificationCodeService struct {
	vcService vcservice.VCService
}

func (s *verificationCodeService) CreateForSigning(cmd *CmdToCreateCodeForSigning) (string, error) {
	p, err := cmd.purpose()
	if err != nil {
		return "", err
	}

	return s.vcService.New(p)
}

func (s *verificationCodeService) ValidateForSigning(cmd *CmdToValidateCodeForSigning) error {
	p, err := cmd.purpose()
	if err != nil {
		return err
	}

	key := domain.NewVerificationCodeKey(cmd.Code, p)

	return s.vcService.Verify(&key)

}

func (s *verificationCodeService) CreateForAddingEmailDomain(cmd *CmdToCreateCodeForEmailDomain) (
	string, error,
) {
	p, err := cmd.purpose()
	if err != nil {
		return "", err
	}

	return s.vcService.New(p)
}

func (s *verificationCodeService) ValidateForAddingEmailDomain(cmd *CmdToValidateCodeForEmailDomain) error {
	p, err := cmd.purpose()
	if err != nil {
		return err
	}

	key := domain.NewVerificationCodeKey(cmd.Code, p)

	return s.vcService.Verify(&key)
}
