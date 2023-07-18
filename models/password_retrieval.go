package models

import "errors"

type PasswordRetrieval struct {
	Password []byte `json:"password"`
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
