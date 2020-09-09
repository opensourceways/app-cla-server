package oauth

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/oauth2"
	"github.com/opensourceways/app-cla-server/util"
)

var clients = map[string]map[string]AuthInterface{}

type AuthInterface interface {
	GetAuthCodeURL(state string) string
	WebRedirectDir() string
	Auth(code, scope string) (string, string, error)
}

func RegisterPlatform(credentialFile string) error {
	cfg := authConfigs{}
	if err := util.LoadFromYaml(credentialFile, &cfg); err != nil {
		return err
	}

	f := func(ac actionConfig, platform string) AuthInterface {
		return &client{
			c:              oauth2.NewOauth2Client(ac.Oauth2Config),
			platform:       platform,
			webRedirectDir: ac.WebRedirectDir,
		}
	}

	for _, item := range cfg.Configs {
		platform := item.Platform

		clients[platform] = map[string]AuthInterface{
			"login": f(item.Login, platform),
			"sign":  f(item.Sign, platform),
		}
	}
	return nil
}

func GetAuthInstance(platform, purpose string) (AuthInterface, error) {
	c, ok := clients[platform]
	if ok {
		if i, ok1 := c[purpose]; ok1 {
			return i, nil
		}
		return nil, fmt.Errorf("Failed to get oauth instance: unknown purpose: %s", purpose)
	}
	return nil, fmt.Errorf("Failed to get oauth instance: unknown platform: %s", platform)
}
