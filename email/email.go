package email

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/util"
	"golang.org/x/oauth2"
)

var EmailAgent = &emailAgent{emailClients: map[string]IEmail{}}

type IEmail interface {
	SendEmail(msg *EmailMessage) error
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
	From       string    `json:"from"`
	To         []string  `json:"to"`
	Subject    string    `json:"subject"`
	Content    string    `json:"content"`
	Attachment string    `json:"attachment"`
	MIME       string    `json:"mime"`
	SendInfo   *SendInfo `json:"send_info"`
}

type SendInfo struct {
	Email         string        `json:"email"`
	Platform      string        `json:"platform"`
	Token         *oauth2.Token `json:"token"`
	AuthorizeCode string        `json:"authorize_code"`
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
