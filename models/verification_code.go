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

func checkVerificationCode(email, code, purpose string) (string, error) {
	vc := dbmodels.VerificationCode{
		Email:   email,
		Code:    code,
		Purpose: purpose,
	}

	err := dbmodels.GetDB().GetVerificationCode(&vc)
	if err == nil {
		if vc.Expiry < util.Now() {
			return util.ErrVerificationCodeExpired, fmt.Errorf("verification code is expired")
		}

		return "", err
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return util.ErrWrongVerificationCode, err
	}
	e := parseDBError(err)
	return e.ErrCode(), e.Err
}
