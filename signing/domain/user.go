package domain

import (
	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

type User struct {
	Id             string
	Account        dp.Account
	Password       []byte // encrypted
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

	logs.Info("login time:%d, period = %d, now=%d", u.LoginTime, config.PeriodOfLoginChecking, now)

	if u.LoginTime+config.PeriodOfLoginChecking < now {
		logs.Info("first time")

		u.LoginTime = now
		u.FailedNum = 1
	} else {
		u.FailedNum += 1

		logs.Info("add one")
		if u.FailedNum >= config.MaxNumOfFailedLogin {
			logs.Info("frozen")

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
