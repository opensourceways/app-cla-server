package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func CreateVerificationCode(email, purpose string, expiry int64) (string, error) {
	code := util.RandStr(6, "number")

	vc := dbmodels.VerificationCode{
		Email:   email,
		Code:    code,
		Purpose: purpose,
		Expiry:  util.Now() + expiry,
	}

	err := dbmodels.GetDB().CreateVerificationCode(vc)
	return code, err
}

func checkVerificationCode(email, code, purpose string) *ModelError {
	vc := dbmodels.VerificationCode{
		Email:   email,
		Code:    code,
		Purpose: purpose,
	}

	err := dbmodels.GetDB().GetVerificationCode(&vc)
	if err == nil {
		if vc.Expiry < util.Now() {
			return newModelError(ErrVerificationCodeExpired, fmt.Errorf("verification code is expired"))
		}

		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrWrongVerificationCode, err)
	}

	return parseDBError(err)
}
