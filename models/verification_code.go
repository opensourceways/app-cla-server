package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type CmdToCreateVerificationCode struct {
	Email   string
	Purpose string
	Expiry  int64
}

func (cmd *CmdToCreateVerificationCode) toCode() dbmodels.VerificationCode {
	return dbmodels.VerificationCode{
		Email:   cmd.Email,
		Code:    util.RandStr(6, "number"),
		Purpose: cmd.Purpose,
		Expiry:  util.Now() + cmd.Expiry,
	}
}

func CreateVerificationCode(cmd CmdToCreateVerificationCode) (string, IModelError) {
	code := cmd.toCode()
	err := dbmodels.GetDB().CreateVerificationCode(&code)
	if err == nil {
		return code.Code, nil
	}
	return code.Code, parseDBError(err)
}

func checkVerificationCode(email, code, purpose string) IModelError {
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
