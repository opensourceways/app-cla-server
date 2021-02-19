package platforms

import (
	"fmt"
)

const (
	errMsgNoPublicEmail          = "no pulic email"
	errMsgRefuseToAuthorizeEmail = "refuse to authorize email"
)

type Platform interface {
	GetUser() (string, error)
	GetAuthorizedEmail() (string, error)
	IsOrgExist(org string) (bool, error)
	ListOrg() ([]string, error)
}

func NewPlatform(accessToken, refreshToken, platform string) (Platform, error) {
	switch platform {
	case "gitee":
		return newGiteeClient(accessToken, refreshToken), nil
	case "github":
		return newGithubClient(accessToken, refreshToken), nil
	}
	return nil, fmt.Errorf("unknown platform:%s", platform)
}

func IsErrOfNoPulicEmail(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == errMsgNoPublicEmail
}

func IsErrOfRefusedToAuthorizeEmail(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == errMsgRefuseToAuthorizeEmail
}
