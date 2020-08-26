package email

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

func getOauth2CodeURL(state string, cfg *oauth2.Config) string {
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func fetchOauth2Token(cfg *oauth2.Config, code string) (*oauth2.Token, error) {
	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token: %v", err)
	}
	return token, nil
}
