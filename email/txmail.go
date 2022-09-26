package email

import (
	"regexp"
	"strings"

	"gopkg.in/gomail.v2"
)

func init() {
	EmailAgent.emailClients["txmail"] = &txmailClient{}
}

type txmailClient struct {
}

var txHost string
var txPort int

func (this *txmailClient) initialize(path string) error {

	txHost = "smtp.exmail.qq.com"
	txPort = 465

	return nil
}

func (this *txmailClient) SendEmail(msg *EmailMessage) error {
	msg.From = msg.SendInfo.Email
	m, err := this.createTxMailMessage(msg)
	if err != nil {
		return err
	}

	d := gomail.NewDialer(txHost, txPort, msg.From, msg.SendInfo.AuthorizeCode)

	err = d.DialAndSend(m)
	if err != nil {
		return err
	}
	return nil
}

func (this *txmailClient) createTxMailMessage(msg *EmailMessage) (*gomail.Message, error) {
	if msg.Attachment == "" {
		return simpleTxmailMessage(msg), nil
	}
	m := gomail.NewMessage()
	m.SetHeader("From", msg.From)
	m.SetHeader("To", msg.To[0])
	m.SetHeader("Subject", msg.Subject)
	m.SetBody("text/plain", msg.Content)
	m.Attach(msg.Attachment)
	return nil, nil
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
