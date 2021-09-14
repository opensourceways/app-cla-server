package models

import (
	"encoding/json"
	"fmt"
)

const passwordRetrievalDesc = "password retrieval"

type PasswordRetrieval struct {
	Password string `json:"password"`
}

func (p PasswordRetrieval) Validate() IModelError {
	return checkPassword(p.Password)
}

func (p PasswordRetrieval) Create(linkID string, key []byte) IModelError {
	k := new(passwordRetrievalKey)
	if err := k.init(key); err != nil {
		return err
	}

	if err := checkVerificationCode(k.Email, k.Code, genDescOfPasswordRetrieval(linkID)); err != nil {
		return err
	}

	m := CorporationManagerResetPassword{
		NewPassword: p.Password,
	}
	return m.Reset(linkID, k.Email)
}

type PasswordRetrievalKey struct {
	Email string `json:"email" required:"true"`
}

func (p PasswordRetrievalKey) Create(linkID string, expiry int64) ([]byte, IModelError) {
	code, mErr := CreateVerificationCode(p.Email, genDescOfPasswordRetrieval(linkID), expiry)
	if mErr != nil {
		return nil, mErr
	}

	k := passwordRetrievalKey{
		Email: p.Email,
		Code:  code,
	}

	return k.encode()
}

func (p PasswordRetrievalKey) Validate() IModelError {
	if p.Email == "" {
		return newModelError(ErrMissgingEmail, fmt.Errorf("missing email"))
	}
	return nil
}

type passwordRetrievalKey struct {
	Email string `json:"email" required:"true"`
	Code  string `json:"code" required:"true"`
}

func (p passwordRetrievalKey) encode() ([]byte, IModelError) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, newModelError(ErrSystemError, err)
	}

	return b, nil
}

func (p *passwordRetrievalKey) init(key []byte) IModelError {
	if err := json.Unmarshal(key, p); err != nil {
		return newModelError(ErrInvalidPWRetrievalKey, err)
	}
	return nil
}

func genDescOfPasswordRetrieval(linkID string) string {
	return passwordRetrievalDesc + linkID
}
