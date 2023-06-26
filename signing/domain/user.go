package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type User struct {
	Account        dp.Account
	Password       dp.Password
	EmailAddr      dp.EmailAddr
	LinkId         string
	CorpSigningId  string
	PasswordChaged bool
}

func (u *User) ChangePassword(p dp.Password) {
	u.Password = p
	u.PasswordChaged = true
}
