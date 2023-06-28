package repository

import "github.com/opensourceways/app-cla-server/signing/domain"

type VerificationCode interface {
	Add(*domain.VerificationCode) error
	Find(*domain.VerificationCodeKey) (domain.VerificationCode, error)
}
