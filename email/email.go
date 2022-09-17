package email

import (
	"fmt"

	"golang.org/x/oauth2"

	"github.com/opensourceways/app-cla-server/util"
)

var EmailAgent = &emailAgent{emailClients: map[string]IEmail{}}

type IEmail interface {
	GetOauth2CodeURL(state string) string
	GetToken(code, scope string) (*oauth2.Token, error)
	GetAuthorizedEmail(token *oauth2.Token) (string, error)
	SendEmail(token *oauth2.Token, Authorize string, msg *EmailMessage) error
	initialize(string) error
}

func Initialize(configFile string) error {
	if err := initTemplate(); err != nil {
		return err
	}

	cfg := emailConfigs{}
	if err := util.LoadFromYaml(configFile, &cfg); err != nil {
		return err
	}

	EmailAgent.webRedirectDir = cfg.webRedirectDirConfig

	for _, item := range cfg.Configs {
		e, err := EmailAgent.GetEmailClient(item.Platform)
		if err != nil {
			return err
		}
		if err = e.initialize(item.Credentials); err != nil {
			return err
		}
	}
	return nil
}

type EmailMessage struct {
	From       string   `json:"from"`
	To         []string `json:"to"`
	Subject    string   `json:"subject"`
	Content    string   `json:"content"`
	Attachment string   `json:"attachment"`
	MIME       string   `json:"mime"`
}

type emailAgent struct {
	emailClients   map[string]IEmail
	webRedirectDir webRedirectDirConfig
}

func (this *emailAgent) WebRedirectDir(success bool) string {
	if success {
		return this.webRedirectDir.WebRedirectDirOnSuccess
	}
	return this.webRedirectDir.WebRedirectDirOnFailure
}

func (this *emailAgent) GetEmailClient(platform string) (IEmail, error) {
	e, ok := this.emailClients[platform]
	if !ok {
		return nil, fmt.Errorf("it only supports gmail platform currently")
	}

	return e, nil
}
