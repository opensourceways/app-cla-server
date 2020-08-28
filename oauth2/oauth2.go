package oauth2

import (
	"context"
	"fmt"

	liboauth2 "golang.org/x/oauth2"
)

var clients = map[string]map[string]Oauth2Interface{}

type Oauth2Interface interface {
	GetToken(code, scope string) (*liboauth2.Token, error)
	GetOauth2CodeURL(state string) string
	WebRedirectDir() string
}

type client struct {
	cfg *liboauth2.Config

	webRedirectDir string
}

func (this *client) GetToken(code, scope string) (*liboauth2.Token, error) {
	return fetchOauth2Token(this.cfg, code)
}

func (this *client) GetOauth2CodeURL(state string) string {
	return getOauth2CodeURL(state, this.cfg)
}

func (this *client) WebRedirectDir() string {
	return this.webRedirectDir
}

func RegisterPlatform(platform, credentialFile string) error {
	cfg, err := loadFromYaml(credentialFile)
	if err != nil {
		return err
	}

	m := map[string]Oauth2Interface{}
	for _, item := range cfg.Config {
		m[item.Purpose] = &client{
			cfg:            buildOauth2Config(item.oauth2Config),
			webRedirectDir: item.WebRedirectDir,
		}
	}
	clients[platform] = m
	return nil
}

func GetOauth2Instance(platform, purpose string) (Oauth2Interface, error) {
	c, ok := clients[platform]
	if ok {
		if i, ok1 := c[purpose]; ok1 {
			return i, nil
		}
		return nil, fmt.Errorf("Failed to get oauth2 instance: unknown purpose: %s", purpose)
	}
	return nil, fmt.Errorf("Failed to get oauth2 instance: unknown platform: %s", platform)
}

func buildOauth2Config(cfg oauth2Config) *liboauth2.Config {
	return &liboauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       cfg.Scope,
		Endpoint: liboauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
		RedirectURL: cfg.RedirectURL,
	}
}

func getOauth2CodeURL(state string, cfg *liboauth2.Config) string {
	return cfg.AuthCodeURL(state, liboauth2.AccessTypeOffline)
}

func fetchOauth2Token(cfg *liboauth2.Config, code string) (*liboauth2.Token, error) {
	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token: %v", err)
	}
	return token, nil
}
