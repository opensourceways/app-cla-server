package app

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

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

// CmdToResetPassword
type CmdToResetPassword struct {
	NewOne    dp.Password
	LinkId    string
	EmailAddr dp.EmailAddr
}
