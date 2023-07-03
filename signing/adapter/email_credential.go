package adapter

import (
	"errors"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/gmailimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/txmailimpl"
)

func NewEmailCredentialAdapter(
	s app.EmailCredentialService,
	repo repository.EmailCredential,
) *emailCredentialAdatper {
	return &emailCredentialAdatper{s: s, repo: repo}
}

type emailCredentialAdatper struct {
	s    app.EmailCredentialService
	repo repository.EmailCredential
}

func (adapter *emailCredentialAdatper) AddTXmailCredential(email, code string) models.IModelError {
	ec, err := txmailimpl.TXmailClient().GenEmailCredential(email, code)
	if err != nil {
		return toModelError(err)
	}

	if err := adapter.s.Add(&ec); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *emailCredentialAdatper) AddGmailCredential(code, scope string) (string, models.IModelError) {
	ec, hasRefreshToken, err := gmailimpl.GmailClient().GenEmailCredential(code, scope)
	if err != nil {
		return "", toModelError(err)
	}

	if !hasRefreshToken {
		if _, err := adapter.repo.Find(ec.Addr); err != nil {
			if commonRepo.IsErrorResourceNotFound(err) {
				return "", models.NewModelError(
					models.ErrNoRefreshToken, errors.New("no refresh token"),
				)
			}

			return "", toModelError(err)
		}

		return ec.Addr.EmailAddr(), nil
	}

	if err := adapter.s.Add(&ec); err != nil {
		return "", toModelError(err)
	}

	return ec.Addr.EmailAddr(), nil
}
