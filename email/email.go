package email

import (
	"fmt"

	"golang.org/x/oauth2"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

var emails = map[string]IEmail{}

type IEmail interface {
	GetOauth2CodeURL(state string) string
	GetAuthorizedEmail(code, scope string) (*models.OrgEmail, error)
	SendEmail(token *oauth2.Token, msg *EmailMessage) error
	WebRedirectDir() string
	initialize(credentials, webRedirectDir string) error
}

func GetEmailClient(platform string) (IEmail, error) {
	e, ok := emails[platform]
	if !ok {
		return nil, fmt.Errorf("it only supports gmail platform currently")
	}

	return e, nil
}

func RegisterPlatform(configFile string) error {
	if err := initTemplate(); err != nil {
		return err
	}

	cfg := emailConfigs{}
	if err := util.LoadFromYaml(configFile, &cfg); err != nil {
		return err
	}

	for _, item := range cfg.Configs {
		e, err := GetEmailClient(item.Platform)
		if err != nil {
			return err
		}
		return e.initialize(item.Credentials, cfg.WebRedirectDir)
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
