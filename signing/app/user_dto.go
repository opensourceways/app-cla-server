package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

// CmdToLogin
type CmdToLogin struct {
	LinkId           string
	Email            dp.EmailAddr
	Account          dp.Account
	Password         dp.Password
	PrivacyConsented bool
}

func (cmd *CmdToLogin) clear() {
	cmd.Password.Clear()
}

// UserBasicInfoDTO
type UserBasicInfoDTO struct {
	Role             string
	UserId           string
	InitialPWChanged bool
}

// UserLoginDTO
type UserLoginDTO struct {
	Role             string
	Email            string
	UserId           string
	CorpName         string
	CorpSigningId    string
	InitialPWChanged bool
	RetryNum         int
}

// CmdToChangePassword
type CmdToChangePassword struct {
	Id     string
	OldOne dp.Password
	NewOne dp.Password
}

func (cmd *CmdToChangePassword) Validate() error {
	if dp.IsSamePassword(cmd.OldOne, cmd.NewOne) {
		return domain.NewDomainError(domain.ErrorCodeUserSamePassword)
	}

	return nil
}

func (cmd *CmdToChangePassword) clear() {
	cmd.OldOne.Clear()
	cmd.NewOne.Clear()
}

// CmdToResetPassword
type CmdToResetPassword struct {
	NewOne dp.Password
	LinkId string
	Key    string
}

func (cmd *CmdToResetPassword) clear() {
	cmd.NewOne.Clear()
}

// CmdToGenKeyForPasswordRetrieval
type CmdToGenKeyForPasswordRetrieval CmdToCreateVerificationCode

func (cmd *CmdToGenKeyForPasswordRetrieval) purpose() (dp.Purpose, error) {
	return (*CmdToCreateVerificationCode)(cmd).genPurpose(vcTypePasswordRetrieval)
}

// resettingPasswordKey
type resettingPasswordKey struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (k *resettingPasswordKey) toEmail() (dp.EmailAddr, error) {
	return dp.NewEmailAddr(k.Email)
}
