package smtpimpl

import (
	"regexp"
	"strings"

	"gopkg.in/gomail.v2"
)

type Config struct {
	Port     int    `json:"port"`
	Host     string `json:"host"`
	Platform string `json:"platform"`
}

func (cfg *Config) SetDefault() {
	if cfg.Host == "" || cfg.Port <= 0 || cfg.Platform == "" {
		cfg.Port = 465
		cfg.Host = "smtp.exmail.qq.com"
		cfg.Platform = "txmail"
	}
}

type smtpImpl struct {
	cfg Config
}

func (impl *smtpImpl) Send(AuthCode []byte, msg *EmailMessage) error {
	m, err := impl.createTxMailMessage(msg)
	if err != nil {
		return err
	}

	d := gomail.NewDialer(impl.cfg.Host, impl.cfg.Port, msg.From, string(AuthCode))

	return d.DialAndSend(m)
}

func (impl *smtpImpl) createTxMailMessage(msg *EmailMessage) (*gomail.Message, error) {
	attachment := msg.Attachment
	if attachment == "" {
		return simpleTxmailMessage(msg), nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", msg.From)
	m.SetHeader("To", msg.To[0])
	m.SetHeader("Subject", msg.Subject)
	m.SetBody("text/plain", msg.Content.String())
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
		m.SetBody(v, msg.Content.String())
	} else {
		m.SetBody("text/plain", msg.Content.String())
	}

	return m
}
