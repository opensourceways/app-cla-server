package models

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/huaweicloud/golangsdk"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type RetrievePW struct {
	LinkID string `json:"link_id" required:"true"`
	Email  string `json:"email" required:"true"`
	Code   string `json:"code" required:"true"`
}

func (r *RetrievePW) Encrypt() (string, IModelError) {
	js, err := json.Marshal(r)
	if err != nil {
		return "", newModelError(ErrEncryptWithRetrievePW, err)
	}
	t, err := r.newEncryption().Encrypt(js)
	if err != nil {
		return "", newModelError(ErrEncryptWithRetrievePW, err)
	}
	return hex.EncodeToString(t), nil
}

func (r *RetrievePW) Decrypt(ciphertext string) IModelError {
	dst, err := hex.DecodeString(ciphertext)
	if err != nil {
		return newModelError(ErrDecryptWithRetrievePW, err)
	}
	s, err := r.newEncryption().Decrypt(dst)
	if err != nil {
		return newModelError(ErrDecryptWithRetrievePW, err)
	}
	if err = json.Unmarshal(s, r); err != nil {
		return newModelError(ErrDecryptWithRetrievePW, err)
	}
	return nil
}

func (r *RetrievePW) Validate() IModelError {
	if _, err := golangsdk.BuildRequestBody(r, ""); err != nil {
		return newModelError(ErrValidateRetrievePW, err)
	}
	return nil
}

func (r *RetrievePW) newEncryption() util.SymmetricEncryption {
	e, _ := util.NewSymmetricEncryption(config.AppConfig.SymmetricEncryptionKey, "")
	return e
}

func CreateVerificationCode(email, purpose string, expiry int64) (string, IModelError) {
	code := util.RandStr(6, "number")

	vc := dbmodels.VerificationCode{
		Email:   email,
		Code:    code,
		Purpose: purpose,
		Expiry:  util.Now() + expiry,
	}

	err := dbmodels.GetDB().CreateVerificationCode(vc)
	if err == nil {
		return code, nil
	}
	return code, parseDBError(err)
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
