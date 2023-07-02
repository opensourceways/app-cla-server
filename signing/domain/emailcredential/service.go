package emailcredential

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/symmetricencryption"
)

func NewEmailCredential(
	repo repository.EmailCredential,
	encrypt symmetricencryption.Encryption,
) *emailCredential {
	return &emailCredential{
		repo:    repo,
		encrypt: encrypt,
	}
}

type EmailCredential interface {
	Add(e *domain.EmailCredential) error
	Find(ed dp.EmailAddr) (domain.EmailCredential, error)
}

type emailCredential struct {
	repo    repository.EmailCredential
	encrypt symmetricencryption.Encryption
}

func (s *emailCredential) Add(e *domain.EmailCredential) error {
	token, err := s.encrypt.Encrypt(e.Token)
	if err != nil {
		return err
	}

	e.Token = token

	return s.repo.Add(e)
}

func (s *emailCredential) Find(ed dp.EmailAddr) (domain.EmailCredential, error) {
	e, err := s.repo.Find(ed)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = domain.NewDomainError(domain.ErrorCodeEmailCredentialNotFound)
		}

		return e, err
	}

	token, err := s.encrypt.Decrypt(e.Token)
	if err == nil {
		e.Token = token
	}

	return e, err
}
