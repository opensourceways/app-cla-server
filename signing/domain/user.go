package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type User struct {
	Account        dp.Account
	EmailAddr      dp.EmailAddr
	Password       dp.Password
	CorpSigningId  string
	PasswordChaged bool
}

func (u *User) ChangePassword(p dp.Password) {
	u.Password = p
	u.PasswordChaged = true
}
