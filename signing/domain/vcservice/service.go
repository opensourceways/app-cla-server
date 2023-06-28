package vcservice

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/randomcode"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

var invalidCode = domain.NewDomainError(domain.ErrorCodeVerificationCodeWrong)

func NewVCService(
	repo repository.VerificationCode,
	randomCode randomcode.RandomCode,
) VCService {
	return &vcService{
		repo:       repo,
		randomCode: randomCode,
	}
}

type VCService interface {
	New(purpose dp.Purpose) (string, error)
	Verify(key *domain.VerificationCodeKey) error
}

type vcService struct {
	repo       repository.VerificationCode
	randomCode randomcode.RandomCode
}

func (s *vcService) Verify(key *domain.VerificationCodeKey) error {
	if !s.randomCode.IsValid(key.Code) {
		return invalidCode
	}

	v, err := s.repo.Find(key)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = invalidCode
		}
		return err
	}

	if v.IsExpired() {
		return invalidCode
	}

	return nil
}

func (s *vcService) New(purpose dp.Purpose) (string, error) {
	code, err := s.randomCode.New()
	if err != nil {
		return "", err
	}

	vc := domain.NewVerificationCode(code, purpose)
	err = s.repo.Add(&vc)

	return code, err
}
