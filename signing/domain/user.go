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
	if !u.IsCorrectPassword(old) {
		return NewDomainError(ErrorCodeUserUnmatchedPassword)
	}

	if u.IsCorrectPassword(newOne) {
		return NewDomainError(ErrorCodeUserSamePassword)
	}

	u.Password = newOne
	u.PasswordChaged = true

	return nil
}

func (u *User) ResetPassword(newOne dp.Password) {
	u.Password = newOne
	u.PasswordChaged = true
}

func (u *User) IsCorrectPassword(p dp.Password) bool {
	return u.Password.Password() == p.Password()
}
