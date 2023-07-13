package platforms

import "fmt"

type Platform interface {
	GetUser() (string, error)
	ListOrg() ([]string, error)
}

func NewPlatform(accessToken, platform string) (Platform, error) {
	switch platform {
	case "gitee":
		return newGiteeClient(accessToken), nil

	case "github":
		return newGithubClient(accessToken), nil
	}

	return nil, fmt.Errorf("unknown platform:%s", platform)
}
