package platforms

import (
	"fmt"

	"golang.org/x/oauth2"
)

type Platform interface {
	GetUser() (string, error)
	ListOrg() ([]string, error)
	GetToken() string
}

func NewPlatform(accessToken, refreshToken, platform string) (Platform, error) {
	switch platform {
	case "gitee":
		return newGiteeClient(accessToken, refreshToken), nil
	}
	return nil, fmt.Errorf("unknown platform:%s", platform)
}

func GetOauthEndpoint(platform string) (oauth2.Endpoint, error) {
	switch platform {
	case "gitee":
		return giteeClient{}.GetOauthEndpoint(), nil
	}
	return oauth2.Endpoint{}, fmt.Errorf("unknown platform:%s", platform)
}
