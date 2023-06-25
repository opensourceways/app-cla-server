package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
)

func NewCorpAdminService(
	repo repository.CorpSigning,
	userService userservice.UserService,
) CorpAdminService {
	return &corpAdminService{
		repo:        repo,
		userService: userService,
	}
}

type CorpAdminService interface {
	Add(string) (err error)
}

type corpAdminService struct {
	repo        repository.CorpSigning
	userService userservice.UserService
}

func (s *corpAdminService) Add(csId string) error {
	cs, err := s.repo.Find(csId)
	if err != nil {
		return err
	}

	if cs.HasAdmin() {
		return domain.NewDomainError(domain.ErrorCodeCorpAdminExists)
	}

	n, err := s.repo.Count(cs.PrimaryEmailDomain())
	if err != nil {
		return err
	}

	if err := cs.SetAdmin(n); err != nil {
		return err
	}

	if err = s.userService.Add(csId, []domain.Manager{cs.Admin}); err != nil {
		return err
	}

	if err = s.repo.AddAdmin(&cs); err != nil {
		s.userService.Remove([]domain.Manager{cs.Admin})
	}

	return nil
}
