package oauth2

import (
	"context"
	"fmt"

	liboauth2 "golang.org/x/oauth2"
)

type Oauth2Interface interface {
	GetToken(code, scope string) (*liboauth2.Token, error)
	GetOauth2CodeURL(state string) string
}

type client struct {
	cfg *liboauth2.Config
}

func (this *client) GetToken(code, scope string) (*liboauth2.Token, error) {
	return FetchOauth2Token(this.cfg, code)
}

func (this *client) GetOauth2CodeURL(state string) string {
	return GetOauth2CodeURL(state, this.cfg)
}

type Oauth2Config struct {
	ClientID     string   `json:"client_id" required:"true"`
	ClientSecret string   `json:"client_secret" required:"true"`
	AuthURL      string   `json:"auth_url" required:"true"`
	TokenURL     string   `json:"token_url" required:"true"`
	RedirectURL  string   `json:"redirect_url" required:"true"`
	Scope        []string `json:"scope" required:"true"`
}

func NewOauth2Client(cfg Oauth2Config) Oauth2Interface {
	return &client{
		cfg: buildOauth2Config(cfg),
	}
}

func buildOauth2Config(cfg Oauth2Config) *liboauth2.Config {
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

func GetOauth2CodeURL(state string, cfg *liboauth2.Config) string {
	return cfg.AuthCodeURL(state, liboauth2.AccessTypeOffline)
}

func FetchOauth2Token(cfg *liboauth2.Config, code string) (*liboauth2.Token, error) {
	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token: %v", err)
	}
	return token, nil
}
