package repositoryimpl

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func NewVerificationCode(dao dao) *verificationCode {
	return &verificationCode{
		dao: dao,
	}
}

type verificationCode struct {
	dao dao
}

func (impl *verificationCode) Add(*domain.VerificationCode) error {
	return nil
}

func (impl *verificationCode) Find(*domain.VerificationCodeKey) (domain.VerificationCode, error) {
	return domain.VerificationCode{}, nil
}
