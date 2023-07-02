package gmailimpl

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
)

const platform = "gmail"

func Platform() string {
	return platform
}

type GetCredential func(dp.EmailAddr) (domain.EmailCredential, error)

type EmailMessage = emailservice.EmailMessage

var gcli = &gmailClient{}

func Init(cfg *Config) error {
	if err := gcli.initialize(cfg.Credentials); err != nil {
		return err
	}

	emailservice.Register(platform, &emailServiceImpl{})

	return nil
}

func GmailClient() *gmailClient {
	return gcli
}

type emailServiceImpl struct {
	getCredential GetCredential
}

func (impl *emailServiceImpl) SendEmail(msg *emailservice.EmailMessage) error {
	// TODO get token
	return gcli.sendEmail(nil, msg)
}
