package oauth2

import (
	"context"
	"fmt"

	liboauth2 "golang.org/x/oauth2"
)

var client = map[string]Oauth2Interface{}

type Oauth2Interface interface {
	GetToken(code, scope, target string) (*liboauth2.Token, error)
	GetOauth2CodeURL(state, target string) (string, error)
	Initialize(credentialFile string) error
}

func RegisterPlatform(platform, credentialFile string) error {
	c, ok := client[platform]
	if !ok {
		return fmt.Errorf("Failed to register a platform: %s is unknown", platform)
	}

	return c.Initialize(credentialFile)
}

func GetOauth2Instance(platform string) Oauth2Interface {
	c, ok := client[platform]
	if ok {
		return c
	}
	return nil
}

type oauth2Config struct {
	ClientID     string   `json:"client_id" required:"true"`
	ClientSecret string   `json:"client_secret" required:"true"`
	AuthURL      string   `json:"auth_url" required:"true"`
	TokenURL     string   `json:"token_url" required:"true"`
	RedirectURL  string   `json:"redirect_url" required:"true"`
	Scope        []string `json:"scope" required:"true"`
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
