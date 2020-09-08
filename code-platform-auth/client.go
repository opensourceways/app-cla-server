package oauth

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/code-platform-auth/platforms"
	"github.com/opensourceways/app-cla-server/oauth2"
)

type client struct {
	c oauth2.Oauth2Interface

	webRedirectDir string

	platform string
}

func (this *client) GetAuthCodeURL(state string) string {
	return this.c.GetOauth2CodeURL(state)
}

func (this *client) WebRedirectDir() string {
	return this.webRedirectDir
}

func (this *client) Auth(code, scope string) (string, string, error) {
	token, err := this.c.GetToken(code, scope)
	if err != nil {
		return "", "", fmt.Errorf("Get token failed: %s", err.Error())
	}

	p, err := platforms.NewPlatform(token.AccessToken, "", this.platform)
	if err != nil {
		return "", "", err
	}

	user, err := p.GetUser()
	if err != nil {
		return "", "", fmt.Errorf("get user failed: %s", err.Error())
	}
	return token.AccessToken, user, nil
}
