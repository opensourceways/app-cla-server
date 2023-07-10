package vcservice

import (
	"time"

	"github.com/beego/beego/v2/core/logs"
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/limiter"
	"github.com/opensourceways/app-cla-server/signing/domain/randomcode"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

var invalidCode = domain.NewDomainError(domain.ErrorCodeVerificationCodeWrong)

func NewVCService(
	repo repository.VerificationCode,
	limiter limiter.Limiter,
	randomCode randomcode.RandomCode,
) VCService {
	return &vcService{
		repo:       repo,
		limiter:    limiter,
		randomCode: randomCode,
	}
}

type VCService interface {
	New(purpose dp.Purpose) (string, error)
	NewIfItCan(purpose dp.Purpose, interval time.Duration) (string, error)
	Verify(key *domain.VerificationCodeKey) error
}

type vcService struct {
	repo       repository.VerificationCode
	limiter    limiter.Limiter
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

func (s *vcService) NewIfItCan(purpose dp.Purpose, interval time.Duration) (string, error) {
	b, err := s.limiter.IsAllowed(purpose.Purpose())
	if err != nil {
		logs.Error("failed to check if it is busy to create code, err:%s", err.Error())
	}

	if !b {
		return "", domain.NewDomainError(domain.ErrorCodeVerificationCodeBusy)
	}

	code, err := s.randomCode.New()
	if err != nil {
		return "", err
	}

	vc := domain.NewVerificationCode(code, purpose)
	if err = s.repo.Add(&vc); err == nil {
		if err1 := s.limiter.Add(purpose.Purpose(), interval); err1 != nil {
			logs.Error("add limiter failed, err:%s", err1.Error())
		}
	}

	return code, err
}
