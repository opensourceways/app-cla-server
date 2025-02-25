package domain

import (
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

type User struct {
	Id              string
	LinkId          string
	Account         dp.Account
	Password        []byte // encrypted
	EmailAddr       dp.EmailAddr
	CorpSigningId   string
	PrivacyConsent  PrivacyConsent
	PasswordChanged bool
	Version         int
}

type PrivacyConsent struct {
	Time    string
	Version string
}

func (u *User) ResetPassword(newOne []byte) {
	u.Password = newOne
	u.PasswordChanged = true
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

func (u *User) UpdatePrivacyConsent(version string) bool {
	if u.PrivacyConsent.Version == version {
		return false
	}

	u.PrivacyConsent = PrivacyConsent{
		Time:    util.Time(),
		Version: version,
	}

	return true
}
