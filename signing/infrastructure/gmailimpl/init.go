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

var gcli = &gmailClient{}

func GmailClient() *gmailClient {
	return gcli
}

func Init(cfg *Config) error {
	if err := gcli.initialize([]byte(cfg.Credentials)); err != nil {
		return err
	}

	emailservice.Register(platform, &emailServiceImpl{})

	return nil
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
