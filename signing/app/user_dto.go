package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

// CmdToLogin
type CmdToLogin struct {
	LinkId   string
	Email    dp.EmailAddr
	Account  dp.Account
	Password dp.Password
}

// UserLoginDTO
type UserLoginDTO struct {
	Role             string
	Email            string
	CorpName         string
	CorpSigningId    string
	InitialPWChanged bool
}

// CmdToChangePassword
type CmdToChangePassword struct {
	Id     string
	OldOne dp.Password
	NewOne dp.Password
}

func (cmd *CmdToChangePassword) Validate() error {
	if cmd.OldOne.Password() == cmd.NewOne.Password() {
		return domain.NewDomainError(domain.ErrorCodeUserSamePassword)
	}

	return nil
}

// CmdToResetPassword
type CmdToResetPassword struct {
	NewOne dp.Password
	LinkId string
	Key    string
}

// CmdToGenKeyForPasswordRetrieval
type CmdToGenKeyForPasswordRetrieval CmdToCreateCodeForSigning

func (cmd *CmdToGenKeyForPasswordRetrieval) purpose() (dp.Purpose, error) {
	return (*CmdToCreateCodeForSigning)(cmd).newPurpose("password retrieval")
}

// resettingPasswordKey
type resettingPasswordKey struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (k *resettingPasswordKey) toEmail() (dp.EmailAddr, error) {
	return dp.NewEmailAddr(k.Email)
}
