package emailservice

import "errors"

var (
	impl = &emailServiceImpl{
		clients: make(map[string]iEmail),
	}
)

func SendEmail(platform string, e *EmailMessage) error {
	return impl.sendEmail(platform, e)
}

func Register(platform string, e iEmail) {
	impl.add(platform, e)
}

type IEmailMessageBulder interface {
	// msg returned only includes content
	GenEmailMsg() (*EmailMessage, error)
}

type EmailMessage struct {
	From       string   `json:"from"`
	To         []string `json:"to"`
	Subject    string   `json:"subject"`
	Content    string   `json:"content"`
	Attachment string   `json:"attachment"`
	MIME       string   `json:"mime"`
}

type iEmail interface {
	SendEmail(msg *EmailMessage) error
}

// emailServiceImpl
type emailServiceImpl struct {
	clients map[string]iEmail
}

func (impl *emailServiceImpl) add(platform string, e iEmail) {
	impl.clients[platform] = e
}

func (impl *emailServiceImpl) sendEmail(platform string, e *EmailMessage) error {
	cli, ok := impl.clients[platform]
	if !ok {
		return errors.New("unsupported email platform")
	}

	return cli.SendEmail(e)
}
