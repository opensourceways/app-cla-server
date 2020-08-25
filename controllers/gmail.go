package controllers

import (
	"context"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"

	"github.com/zengchen1024/cla-server/models"
)

type gmailInfo struct{}

func (this gmailInfo) GetOauth2Config(path string) (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, this.getScope()...)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	return config, nil
}

func (this gmailInfo) getScope() []string {
	return []string{gmail.GmailReadonlyScope, gmail.GmailSendScope}
}

func (this gmailInfo) GenOrgEmail(code, path, scope string) (*models.OrgEmail, error) {
	config, err := this.GetOauth2Config(path)
	if err != nil {
		return nil, err
	}

	token, err := fetchOauth2Token(config, code)
	if err != nil {
		return nil, err
	}

	client := config.Client(context.Background(), token)
	srv, err := gmail.New(client)
	if err != nil {
		return nil, err
	}

	v, err := srv.Users.GetProfile("me").Do()
	if err != nil {
		return nil, err
	}

	return &models.OrgEmail{
		Email: v.EmailAddress,
		Token: token,
	}, nil
}
