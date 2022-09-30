package email

import (
	"regexp"
	"strings"

	"gopkg.in/gomail.v2"
)

func init() {
	EmailAgent.emailsAuthedByCode["txmail"] = &txmailClient{}
}

type txmailConfig struct {
	host string
	port int
}

type txmailClient struct {
	cfg *txmailConfig
}

func (this *txmailClient) initialize(path string) error {
	this.cfg = &txmailConfig{
		host: "smtp.exmail.qq.com",
		port: 465,
	}

	return nil
}

func (this *txmailClient) SendEmail(AuthCode string, msg *EmailMessage) error {
	m, err := this.createTxMailMessage(msg)
	if err != nil {
		return err
	}

	d := gomail.NewDialer(this.cfg.host, this.cfg.port, msg.From, AuthCode)

	err = d.DialAndSend(m)
	if err != nil {
		return err
	}
	return nil
}

func (this *txmailClient) createTxMailMessage(msg *EmailMessage) (*gomail.Message, error) {
	attachment := msg.Attachment
	if attachment == "" {
		return simpleTxmailMessage(msg), nil
	}
	m := gomail.NewMessage()
	m.SetHeader("From", msg.From)
	m.SetHeader("To", msg.To[0])
	m.SetHeader("Subject", msg.Subject)
	m.SetBody("text/plain", msg.Content)
	m.Attach(msg.Attachment)
	return m, nil
}

func simpleTxmailMessage(msg *EmailMessage) *gomail.Message {
	mime := make(map[string]string)
	m := gomail.NewMessage()
	m.SetHeader("From", msg.From)
	m.SetHeader("To", msg.To...)
	m.SetHeader("Subject", msg.Subject)
	if msg.MIME != "" {
		reg := regexp.MustCompile("\\s+")
		s := reg.ReplaceAllString(msg.MIME, "")
		he := strings.Split(s, ";")
		for _, v := range he {
			de := strings.Split(v, ":")
			if len(de) >= 2 {
				mime[de[0]] = de[1]
			}
		}
	}
	if v, ok := mime["Content-Type"]; ok {
		m.SetBody(v, msg.Content)
	} else {
		m.SetBody("text/plain", msg.Content)
	}
	return m
}
