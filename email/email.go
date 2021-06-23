package email

import (
	"fmt"

	"golang.org/x/oauth2"

	"github.com/opensourceways/app-cla-server/util"
)

var EmailAgent = &emailAgent{emailClients: map[string]IEmail{}}

type IEmail interface {
	SwitchAuthType(state string, way string) string
	GetToken(code, scope string) (*oauth2.Token, error)
	GetAuthorizedEmail(token *oauth2.Token) (string, error)
	SendEmail(token *oauth2.Token, msg *EmailMessage) error
	initialize(platformConfig) error
}

func Initialize(configFile string) error {
	if err := initTemplate(); err != nil {
		return err
	}

	cfg := emailConfigs{}
	if err := util.LoadFromYaml(configFile, &cfg); err != nil {
		return err
	}

	if err := cfg.validate(); err != nil {
		return err
	}

	EmailAgent.cfg = cfg

	for _, item := range cfg.Configs {
		e, err := EmailAgent.GetEmailClient(item.Platform)
		if err != nil {
			return err
		}
		if err = e.initialize(item); err != nil {
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
}

type emailAgent struct {
	emailClients map[string]IEmail
	cfg          emailConfigs
}

func (this *emailAgent) WebRedirectDir(success bool, way string) string {
	return this.cfg.redirectWebDir(success, way)
}

func (this *emailAgent) GetEmailClient(platform string) (IEmail, error) {
	e, ok := this.emailClients[platform]
	if !ok {
		return nil, fmt.Errorf("it only supports gmail platform currently")
	}

	return e, nil
}
