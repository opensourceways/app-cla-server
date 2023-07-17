package platforms

import "fmt"

type Platform interface {
	GetUser(string) (string, error)
	ListOrg(string) ([]string, error)
}

func NewPlatform(platform string) (Platform, error) {
	switch platform {
	case "gitee":
		return newGiteeClient(), nil

	case "github":
		return newGithubClient(), nil
	}

	return nil, fmt.Errorf("unknown platform:%s", platform)
}
