package smtpimpl

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
)

var smtp *smtpImpl

func SMTP() *smtpImpl {
	return smtp
}

func Init(cfg *Config) {
	smtp = &smtpImpl{
		cfg: *cfg,
	}
}

func Platform() string {
	return smtp.cfg.Platform
}

func RegisterEmailService(f GetCredential) {
	emailservice.Register(Platform(), &emailServiceImpl{f})
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

	logs.Info("smtp code = %#v, %v, %v, %v", c, c.Addr.EmailAddr(), c.Platform, c.Token)

	err = smtp.Send(c.Token, msg)
	c.Clear()

	return err
}
