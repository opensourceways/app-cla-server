package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type User struct {
	Id             string
	Account        dp.Account
	Password       []byte // encrypted
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

func (u *User) ChangePassword(
	isCorrect func([]byte) bool,
	genNewPassword func() ([]byte, error),
) error {
	if !isCorrect(u.Password) {
		return NewDomainError(ErrorCodeUserUnmatchedPassword)
	}

	v, err := genNewPassword()
	if err != nil {
		return err
	}

	u.ResetPassword(v)

	return nil
}
