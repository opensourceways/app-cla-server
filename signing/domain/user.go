package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type User struct {
	Id             string
	Account        dp.Account
	Password       []byte
	EmailAddr      dp.EmailAddr
	LinkId         string
	CorpSigningId  string
	PasswordChaged bool
	Version        int
}

func (u *User) ResetPassword(newOne []byte) {
	u.Password = newOne
	u.PasswordChaged = true
}

func (u *User) UserPassword() []byte {
	return u.Password
}
