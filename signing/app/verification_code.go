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

func NewSigningCodeService(
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
	return s.vcService.New(cmd.purpose())
}

func (s *verificationCodeService) ValidateForSigning(cmd *CmdToValidateCodeForSigning) error {
	key := domain.NewVerificationCodeKey(cmd.Code, cmd.purpose())

	return s.vcService.Verify(&key)

}

func (s *verificationCodeService) CreateForAddingEmailDomain(cmd *CmdToCreateCodeForEmailDomain) (
	string, error,
) {
	return s.vcService.New(cmd.purpose())
}

func (s *verificationCodeService) ValidateForAddingEmailDomain(cmd *CmdToValidateCodeForEmailDomain) error {
	key := domain.NewVerificationCodeKey(cmd.Code, cmd.purpose())

	return s.vcService.Verify(&key)
}
