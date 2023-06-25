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
	Add(string) (ManagerDTO, error)
}

type corpAdminService struct {
	repo        repository.CorpSigning
	userService userservice.UserService
}

func (s *corpAdminService) Add(csId string) (dto ManagerDTO, err error) {
	cs, err := s.repo.Find(csId)
	if err != nil {
		return
	}

	if cs.HasAdmin() {
		err = domain.NewDomainError(domain.ErrorCodeCorpAdminExists)

		return
	}

	n, err := s.repo.Count(cs.Link.Id, cs.PrimaryEmailDomain())
	if err != nil {
		return
	}

	if err = cs.SetAdmin(n); err != nil {
		return
	}

	pws, err := s.userService.Add(csId, []domain.Manager{cs.Admin})
	if err != nil {
		return
	}

	if err = s.repo.AddAdmin(&cs); err != nil {
		s.userService.Remove([]domain.Manager{cs.Admin})

		return
	}

	admin := &cs.Admin
	dto = ManagerDTO{
		Id:        admin.Id,
		Name:      admin.Name.Name(),
		Password:  pws[admin.Id],
		EmailAddr: admin.EmailAddr.EmailAddr(),
	}

	return
}
