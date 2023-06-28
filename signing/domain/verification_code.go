package domain

import (
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

func NewVerificationCode(code string, Purpose dp.Purpose) VerificationCode {
	return VerificationCode{
		Expiry: util.Now() + config.VerificationCodeExpiry,
		VerificationCodeKey: VerificationCodeKey{
			Code:    code,
			Purpose: Purpose,
		},
	}
}

func NewVerificationCodeKey(code string, Purpose dp.Purpose) VerificationCodeKey {
	return VerificationCodeKey{
		Code:    code,
		Purpose: Purpose,
	}
}

type VerificationCodeKey struct {
	Code    string
	Purpose dp.Purpose
}

type VerificationCode struct {
	VerificationCodeKey

	Expiry int64
}

func (vc *VerificationCode) IsExpired() bool {
	return vc.Expiry < util.Now()
}
