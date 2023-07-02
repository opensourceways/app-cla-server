package txmailimpl

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
)

const platform = "txmail"

func Platform() string {
	return platform
}

type GetCredential func(dp.EmailAddr) (domain.EmailCredential, error)

type EmailMessage = emailservice.EmailMessage

var txcli = &txmailClient{}

func Init() {
	txcli.initialize()

	emailservice.Register(platform, &emailServiceImpl{})
}

func TXmailClient() *txmailClient {
	return txcli
}

type emailServiceImpl struct {
	getCredential GetCredential
}

func (impl *emailServiceImpl) SendEmail(msg *emailservice.EmailMessage) error {
	// TODO get code
	return txcli.Send("", msg)
}
