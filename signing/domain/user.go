package domain

import (
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

const communityLink = "link_place_holder"

type User struct {
	LinkId        string
	CorpSigningId string

	UserBasicInfo
}

func (u *User) IsCommunityManager() bool {
	return u.LinkId == communityLink
}

func (u *User) CommunityManagerLinkId() string {
	return communityLink
}

type PrivacyConsent struct {
	Time    string
	Version string
}

type UserBasicInfo struct {
	Id              string
	Account         dp.Account
	Password        []byte // encrypted
	EmailAddr       dp.EmailAddr
	PrivacyConsent  PrivacyConsent
	PasswordChanged bool
	Version         int
}

func (u *UserBasicInfo) ResetPassword(newOne []byte) {
	u.Password = newOne
	u.PasswordChanged = true
}

func (u *UserBasicInfo) ChangePassword(
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

func (u *UserBasicInfo) UpdatePrivacyConsent(version string) bool {
	if u.PrivacyConsent.Version == version {
		return false
	}

	u.PrivacyConsent = PrivacyConsent{
		Time:    util.Time(),
		Version: version,
	}

	return true
}
