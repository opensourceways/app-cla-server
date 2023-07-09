package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/emailcredential"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/gmailimpl"
)

func NewGmailService(
	repo repository.EmailCredential,
	es emailcredential.EmailCredential,
) *gmailService {
	return &gmailService{repo: repo, es: es}
}

type GmailService interface {
	Authorize(cmd *CmdToAuthorizeGmail) (string, error)
}

type CmdToAuthorizeGmail struct {
	Code  string
	Scope string
}

type gmailService struct {
	es   emailcredential.EmailCredential
	repo repository.EmailCredential
}

func (adapter *gmailService) Authorize(cmd *CmdToAuthorizeGmail) (string, error) {
	ec, hasRefreshToken, err := gmailimpl.GmailClient().GenEmailCredential(
		cmd.Code, cmd.Scope,
	)
	if err != nil {
		return "", err
	}

	if !hasRefreshToken {
		if _, err := adapter.repo.Find(ec.Addr); err != nil {
			if commonRepo.IsErrorResourceNotFound(err) {
				return "", domain.NewDomainError(domain.ErrorCodeGmailNoRefreshToken)
			}

			return "", err
		}

		return ec.Addr.EmailAddr(), nil
	}

	if err := adapter.es.Add(&ec); err != nil {
		return "", err
	}

	return ec.Addr.EmailAddr(), nil
}
