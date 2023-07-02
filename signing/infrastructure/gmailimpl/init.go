package gmailimpl

import (
	"encoding/json"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
)

const platform = "gmail"

var gcli = &gmailClient{}

func GmailClient() *gmailClient {
	return gcli
}

func (cli *gmailClient) GenEmailCredential(code, scope string) (ec domain.EmailCredential, hasRefreshToken bool, err error) {
	token, err := cli.GetToken(code, scope)
	if err != nil {
		return
	}

	emailAddr, err := cli.GetAuthorizedEmail(token)
	if err != nil {
		return
	}

	ec.Platform = platform
	if ec.Addr, err = dp.NewEmailAddr(emailAddr); err != nil {
		return
	}

	ec.Token, err = json.Marshal(token)

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
	// TODO get token
	return gcli.sendEmail(nil, msg)
}
