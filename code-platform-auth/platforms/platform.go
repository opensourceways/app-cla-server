package platforms

import (
	"fmt"
)

type Platform interface {
	GetUser() (string, error)
	IsOrgExist(org string) (bool, error)
}

func NewPlatform(accessToken, refreshToken, platform string) (Platform, error) {
	switch platform {
	case "gitee":
		return newGiteeClient(accessToken, refreshToken), nil
	}
	return nil, fmt.Errorf("unknown platform:%s", platform)
}
