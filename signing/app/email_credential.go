package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/emailcredential"
)

type CmdToAddEmailCredential = domain.EmailCredential

func NewEmailCredentialService(
	es emailcredential.EmailCredential,
) EmailCredentialService {
	return &emailCredentialService{
		es: es,
	}
}

type EmailCredentialService interface {
	Add(cmd *CmdToAddEmailCredential) error
	Find(ed dp.EmailAddr) (domain.EmailCredential, error)
}

type emailCredentialService struct {
	es emailcredential.EmailCredential
}

func (s *emailCredentialService) Add(cmd *CmdToAddEmailCredential) error {
	return s.es.Add(cmd)
}

func (s *emailCredentialService) Find(ed dp.EmailAddr) (domain.EmailCredential, error) {
	return s.es.Find(ed)
}
