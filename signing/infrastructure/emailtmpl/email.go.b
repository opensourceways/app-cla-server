package email

import (
	"errors"

	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
)

var (
	EmailClient = &emailServiceImpl{
		clients: make(map[string]iEmail),
	}
)

type iEmail interface {
	SendEmail(msg *emailservice.EmailMessage) error
}

func Init() error {
	return initTemplate()
}

func Register(platform string, e iEmail) {
	EmailClient.add(platform, e)
}

type emailServiceImpl struct {
	clients map[string]iEmail
}

func (impl *emailServiceImpl) add(platform string, e iEmail) {
	impl.clients[platform] = e
}

func (impl *emailServiceImpl) SendEmail(platform string, e *emailservice.EmailMessage) error {
	cli, ok := impl.clients[platform]
	if !ok {
		return errors.New("unsupported email platform")
	}

	return cli.SendEmail(e)
}
