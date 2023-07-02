package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/emailcredential"
)

type CmdToAddEmailCredential = domain.EmailCredential

func NewEmailCredential(
	es emailcredential.EmailCredential,
) EmailCredential {
	return &emailCredential{
		es: es,
	}
}

type EmailCredential interface {
	Add(cmd *CmdToAddEmailCredential) error
	Find(ed dp.EmailAddr) (domain.EmailCredential, error)
}

type emailCredential struct {
	es emailcredential.EmailCredential
}

func (s *emailCredential) Add(cmd *CmdToAddEmailCredential) error {
	return s.es.Add(cmd)
}

func (s *emailCredential) Find(ed dp.EmailAddr) (domain.EmailCredential, error) {
	return s.es.Find(ed)
}
