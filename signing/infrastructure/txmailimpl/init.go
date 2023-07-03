package txmailimpl

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
)

const platform = "txmail"

var txcli = &txmailClient{}

func TXmailClient() *txmailClient {
	return txcli
}

func (cli *txmailClient) GenEmailCredential(email, code string) (ec domain.EmailCredential, err error) {
	if ec.Addr, err = dp.NewEmailAddr(email); err != nil {
		return
	}

	ec.Token = []byte(code)
	ec.Platform = platform

	return
}

func Init() {
	txcli.initialize()
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

	return txcli.Send(string(c.Token), msg)
}
