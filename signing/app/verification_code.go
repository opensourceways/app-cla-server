package app

import (
	"fmt"
	"time"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

// verificationCodeService
type verificationCodeService struct {
	vc vcservice.VCService
}

func (s *verificationCodeService) newCode(cmd vcPurpose) (string, error) {
	p, err := cmd.purpose()
	if err != nil {
		return "", err
	}

	return s.vc.New(p)
}

func (s *verificationCodeService) newCodeIfItCan(cmd vcPurpose, interval time.Duration) (string, error) {
	p, err := cmd.purpose()
	if err != nil {
		fmt.Printf("newCodeIfItCan, 1")
		return "", err
	}

	v, err := s.vc.NewIfItCan(p, interval)
	fmt.Printf("newCodeIfItCan, 2")

	return v, err
}

func (s *verificationCodeService) validate(cmd vcPurpose, code string) error {
	p, err := cmd.purpose()
	if err != nil {
		return err
	}

	key := domain.NewVerificationCodeKey(code, p)

	return s.vc.Verify(&key)
}
