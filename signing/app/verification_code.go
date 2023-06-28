package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

type VerificationCodeService interface {
	New(vcPurpose) (string, error)
	Validate(vcPurpose, string) error
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

func (s *verificationCodeService) New(cmd vcPurpose) (string, error) {
	p, err := cmd.purpose()
	if err != nil {
		return "", err
	}

	return s.vcService.New(p)
}

func (s *verificationCodeService) Validate(cmd vcPurpose, code string) error {
	p, err := cmd.purpose()
	if err != nil {
		return err
	}

	key := domain.NewVerificationCodeKey(code, p)

	return s.vcService.Verify(&key)
}
