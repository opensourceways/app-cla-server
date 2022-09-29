package email

import (
	"errors"

	"golang.org/x/oauth2"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

var (
	EmailAgent = &emailAgent{
		emailsOauthed:      make(map[string]iEmailOauthed),
		emailsAuthedByCode: make(map[string]iEmailAuthedByCode),
	}

	errorUnsupportedEmail = errors.New("unsupported email platform")
)

type IEmail interface {
	SendEmail(msg *EmailMessage) error
}

type initialize interface {
	initialize(string) error
}

type iEmailOauthed interface {
	GetOauth2CodeURL(state string) string
	GetToken(code, scope string) (*oauth2.Token, error)
	GetAuthorizedEmail(token *oauth2.Token) (string, error)
	SendEmail(token *oauth2.Token, msg *EmailMessage) error

	initialize
}

type iEmailAuthedByCode interface {
	SendEmail(string, *EmailMessage) error

	initialize
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
		if err := EmailAgent.initialize(&item); err != nil {
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
	webRedirectDir webRedirectDirConfig

	emailsOauthed      map[string]iEmailOauthed
	emailsAuthedByCode map[string]iEmailAuthedByCode
}

func (this *emailAgent) WebRedirectDir(success bool) string {
	if success {
		return this.webRedirectDir.WebRedirectDirOnSuccess
	}
	return this.webRedirectDir.WebRedirectDirOnFailure
}

func (ea *emailAgent) initialize(cfg *platformConfig) error {
	if e, ok := ea.emailsOauthed[cfg.Platform]; ok {
		return e.initialize(cfg.Credentials)
	}

	if e, ok := ea.emailsAuthedByCode[cfg.Platform]; ok {
		return e.initialize(cfg.Credentials)
	}

	return errorUnsupportedEmail
}

func (this *emailAgent) GetEmailClient(cfg *models.OrgEmail) (IEmail, error) {
	if e, ok := this.emailsOauthed[cfg.Platform]; ok {
		return emailOauthedAdapter{e: e, token: cfg.Token}, nil
	}

	if e, ok := this.emailsAuthedByCode[cfg.Platform]; ok {
		return emailAuthedByCodeAdapter{e: e, code: cfg.AuthCode}, nil
	}

	return nil, errorUnsupportedEmail
}

func (ea *emailAgent) GetEmailOauthedClient(platform string) (iEmailOauthed, error) {
	e, ok := ea.emailsOauthed[platform]
	if !ok {
		return nil, errorUnsupportedEmail
	}

	return e, nil
}

func (ea *emailAgent) GetEmailAuthClient(platform string) (iEmailAuthedByCode, error) {
	e, ok := ea.emailsAuthedByCode[platform]
	if !ok {
		return nil, errorUnsupportedEmail
	}

	return e, nil
}

type emailOauthedAdapter struct {
	e     iEmailOauthed
	token *oauth2.Token
}

func (ea emailOauthedAdapter) SendEmail(msg *EmailMessage) error {
	return ea.e.SendEmail(ea.token, msg)
}

type emailAuthedByCodeAdapter struct {
	e    iEmailAuthedByCode
	code string
}

func (ea emailAuthedByCodeAdapter) SendEmail(msg *EmailMessage) error {
	return ea.e.SendEmail(ea.code, msg)
}
