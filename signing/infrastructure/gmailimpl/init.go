package gmailimpl

import (
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
)

const platform = "gmail"

var gcli = &gmailClient{}

func GmailClient() *gmailClient {
	return gcli
}

func (cli *gmailClient) GenEmailCredential(code, scope string) (
	ec domain.EmailCredential, hasRefreshToken bool, err error,
) {
	token, err := cli.GetToken(code, scope)
	if err != nil {
		return
	}

	if ec.Token, err = json.Marshal(token); err != nil {
		return
	}

	emailAddr, err := cli.GetAuthorizedEmail(token)
	if err != nil {
		return
	}

	if ec.Addr, err = dp.NewEmailAddr(emailAddr); err != nil {
		return
	}

	ec.Platform = platform

	hasRefreshToken = token.RefreshToken != ""

	return
}

func Init(cfg *Config) error {
	return gcli.initialize([]byte(cfg.Credentials))
}

func RegisterEmailService(f GetCredential) {
	emailservice.Register(platform, &emailServiceImpl{f})
}

type GetCredential func(dp.EmailAddr) (domain.EmailCredential, error)

type EmailMessage = emailservice.EmailMessage

// emailServiceImpl
type emailServiceImpl struct {
	getCredential GetCredential
}

func (impl *emailServiceImpl) SendEmail(msg *emailservice.EmailMessage) error {
	e, err := dp.NewEmailAddr(msg.From)
	if err != nil {
		return err
	}

	c, err := impl.getCredential(e)
	if err != nil {
		return err
	}

	var token oauth2.Token

	if err := json.Unmarshal(c.Token, &token); err != nil {
		return fmt.Errorf("Failed to unmarshal oauth2 token: %s", err.Error())
	}

	return gcli.sendEmail(&token, msg)
}
