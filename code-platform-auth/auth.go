package oauth

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/oauth2"
)

const (
	AuthApplyToLogin = "login"
)

// key is the purpose that authorization applies to
var Auth = map[string]*codePlatformAuth{}

func Initialize(cfg *Config) error {
	f := func(purpose string, ac *authConfig) {
		cpa := &codePlatformAuth{
			webRedirectDir: ac.webRedirectDirConfig,
			clients:        map[string]AuthInterface{},
		}

		for _, item := range ac.Configs {
			cpa.clients[item.Platform] = &authClient{
				c: oauth2.NewOauth2Client(item.Oauth2Config),
			}
		}

		Auth[purpose] = cpa
	}

	f(AuthApplyToLogin, &cfg.Login)
	return nil
}

type AuthInterface interface {
	GetAuthCodeURL(state string) string
	GetToken(code, scope string) (string, error)
}

// codePlatformAuth
type codePlatformAuth struct {
	webRedirectDir webRedirectDirConfig
	clients        map[string]AuthInterface
}

func (auth *codePlatformAuth) GetAuthInstance(platform string) (AuthInterface, error) {
	if c, ok := auth.clients[platform]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("Failed to get oauth instance: unknown platform: %s", platform)
}

func (auth *codePlatformAuth) WebRedirectDir(success bool) string {
	if success {
		return auth.webRedirectDir.WebRedirectDirOnSuccess
	}
	return auth.webRedirectDir.WebRedirectDirOnFailure
}

// authClient
type authClient struct {
	c oauth2.Oauth2Interface
}

func (cli *authClient) GetAuthCodeURL(state string) string {
	return cli.c.GetOauth2CodeURL(state)
}

func (cli *authClient) GetToken(code, scope string) (string, error) {
	token, err := cli.c.GetToken(code, scope)
	if err != nil {
		return "", fmt.Errorf("Get token failed: %s", err.Error())
	}

	return token.AccessToken, nil
}
