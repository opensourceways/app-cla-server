package domain

import (
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

type User struct {
	Id             string
	Account        dp.Account
	Password       []byte
	EmailAddr      dp.EmailAddr
	LinkId         string
	CorpSigningId  string
	PasswordChaged bool
	Version        int

	FrozenTime int64
	LoginTime  int64
	FailedNum  int
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

func (u *User) Login(isCorrect func([]byte) bool) (bool, error) {
	now := util.Now()

	if now < u.FrozenTime {
		return false, NewDomainError(ErrorCodeUserFrozen)
	}

	if isCorrect(u.Password) {
		return false, nil
	}

	if u.LoginTime+config.PeriodOfLoginChecking < now {
		u.LoginTime = now
		u.FailedNum = 1
	} else {
		u.FailedNum += 1

		if u.FailedNum >= config.MaxNumOfFailedLogin {
			u.FrozenTime = now + config.PeriodOfLoginFrozen

			return true, NewDomainError(ErrorCodeUserFrozen)
		}
	}

	return true, NewDomainError(ErrorCodeUserWrongAccountOrPassword)
}

func (u *User) Logout() {
	u.FrozenTime = 0
	u.LoginTime = 0
}
