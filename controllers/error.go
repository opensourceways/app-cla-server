package controllers

import "github.com/opensourceways/app-cla-server/dbmodels"

const (
	ErrorStart = iota
	ErrInvalidParameter
	ErrHasSigned
	ErrHasNotSigned
	ErrMissingToken
	ErrUnknownToken
	ErrInvalidToken
	ErrSigningUncompleted
	ErrUnknownEmailPlatform
	ErrSendingEmail
	ErrWrongVerificationCode
	ErrVerificationCodeExpired
)

func convertDBError(err error) (int, int) {
	e, ok := dbmodels.IsDBError(err)
	if !ok {
		return 500, 0
	}

	switch e.ErrCode {
	case dbmodels.ErrInvalidParameter:
		return 400, ErrInvalidParameter

	case dbmodels.ErrHasSigned:
		return 400, ErrHasSigned

	case dbmodels.ErrHasNotSigned:
		return 400, ErrHasNotSigned

	case dbmodels.ErrWrongVerificationCode:
		return 400, ErrWrongVerificationCode

	case dbmodels.ErrVerificationCodeExpired:
		return 400, ErrVerificationCodeExpired
	}

	return 500, 0
}
