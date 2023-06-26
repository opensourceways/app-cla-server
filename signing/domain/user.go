package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type User struct {
	Id             string
	Account        dp.Account
	Password       dp.Password
	EmailAddr      dp.EmailAddr
	LinkId         string
	CorpSigningId  string
	PasswordChaged bool
	Version        int
}

func (u *User) ChangePassword(old, newOne dp.Password) error {
	if u.Password.Password() != old.Password() {
		return NewDomainError(ErrorCodeUserUnmatchedPassword)
	}

	if u.Password.Password() == newOne.Password() {
		return NewDomainError(ErrorCodeUserSamePassword)
	}

	u.Password = newOne
	u.PasswordChaged = true

	return nil
}
