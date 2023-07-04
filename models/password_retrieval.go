package models

import "errors"

type PasswordRetrieval struct {
	Password string `json:"password"`
}

type PasswordRetrievalKey struct {
	Email string `json:"email" required:"true"`
}

func (p PasswordRetrievalKey) Validate() IModelError {
	if p.Email == "" {
		return newModelError(ErrMissgingEmail, errors.New("missing email"))
	}
	return nil
}

func GenKeyForPasswordRetrieval(linkId string, opt *PasswordRetrievalKey) (string, IModelError) {
	return userAdapterInstance.GenKeyForPasswordRetrieval(
		linkId, opt.Email,
	)
}

func ResetPassword(linkId string, opt *PasswordRetrieval, key string) IModelError {
	return userAdapterInstance.ResetPassword(linkId, key, opt.Password)
}
