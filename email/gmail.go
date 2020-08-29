package email

import (
	"context"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"

	"github.com/zengchen1024/cla-server/models"
	myoauth2 "github.com/zengchen1024/cla-server/oauth2"
)

func init() {
	emails["gmail"] = &gmailClient{}
}

type gmailClient struct {
	cfg *oauth2.Config

	webRedirectDir string
}

func (this *gmailClient) initialize(path, webRedirectDir string) error {
	cfg, err := this.getOauth2Config(path)
	if err != nil {
		return err
	}

	this.cfg = cfg
	this.webRedirectDir = webRedirectDir
	return nil
}

func (this *gmailClient) WebRedirectDir() string {
	return this.webRedirectDir
}

func (this *gmailClient) GetAuthorizedEmail(code, scope string) (*models.OrgEmail, error) {
	if this.cfg == nil {
		return nil, fmt.Errorf("gmail has not been initialized")
	}

	token, err := myoauth2.FetchOauth2Token(this.cfg, code)
	if err != nil {
		return nil, err
	}

	client := this.cfg.Client(context.Background(), token)
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

func (this *gmailClient) GetOauth2CodeURL(state string) string {
	return myoauth2.GetOauth2CodeURL(state, this.cfg)
}

func (this *gmailClient) SendEmail(token oauth2.Token) error {
	client := this.cfg.Client(context.Background(), &token)
	srv, err := gmail.New(client)
	if err != nil {
		return err
	}

	_, err = srv.Users.Messages.Send("me", nil).Do()

	return err
}
func (this *gmailClient) getOauth2Config(path string) (*oauth2.Config, error) {
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

func (this *gmailClient) getScope() []string {
	return []string{gmail.GmailReadonlyScope, gmail.GmailSendScope}
}
