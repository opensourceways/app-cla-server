package oauth

import (
	"fmt"

	"github.com/zengchen1024/cla-server/oauth2"
)

var clients = map[string]map[string]AuthInterface{}

type AuthInterface interface {
	GetAuthCodeURL(state string) string
	WebRedirectDir() string
	Auth(code, scope string) (string, string, error)
}

func RegisterPlatform(platform, credentialFile string) error {
	cfg, err := loadFromYaml(credentialFile)
	if err != nil {
		return err
	}

	m1 := map[string]authConfig{
		"login": cfg.Login,
		"sign":  cfg.Sign,
	}

	m := map[string]AuthInterface{}
	for k, item := range m1 {
		m[k] = &client{
			c:              oauth2.NewOauth2Client(item.Oauth2Config),
			platform:       platform,
			webRedirectDir: item.WebRedirectDir,
		}
	}

	clients[platform] = m
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
