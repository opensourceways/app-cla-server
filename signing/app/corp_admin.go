package app

import (
	"strconv"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
)

func NewCorpAdminService(
	repo repository.CorpSigning,
	linkRepo repository.Link,
	userService userservice.UserService,
) CorpAdminService {
	return &corpAdminService{
		repo:        repo,
		linkRepo:    linkRepo,
		userService: userService,
	}
}

type CorpAdminService interface {
	Add(userId, csId string) (string, ManagerDTO, error)
}

type corpAdminService struct {
	repo        repository.CorpSigning
	linkRepo    repository.Link
	userService userservice.UserService
}

func (s *corpAdminService) Add(userId, csId string) (linkId string, dto ManagerDTO, err error) {
	cs, err := s.repo.Find(csId)
	if err != nil {
		return
	}

	linkId = cs.Link.Id

	if err = checkIfCommunityManager(userId, linkId, s.linkRepo); err != nil {
		return
	}

	if err = cs.CanSetAdmin(); err != nil {
		return
	}

	adminId, err := s.getAdminId(&cs)
	if err != nil {
		return
	}

	cs.SetAdmin(adminId)

	pws, ids, err := s.userService.Add(cs.Link.Id, csId, []domain.Manager{cs.Admin})
	if err != nil {
		return
	}

	if err = s.repo.AddAdmin(&cs); err != nil {
		s.userService.Remove(ids)

		return
	}

	account, err := cs.Admin.Account()
	if err != nil {
		return
	}

	admin := &cs.Admin
	dto = ManagerDTO{
		Role:      domain.RoleAdmin,
		Name:      admin.Name.Name(),
		Account:   account.Account(),
		Password:  pws[admin.Id].Password(),
		EmailAddr: admin.EmailAddr.EmailAddr(),
	}

	return
}

func (s *corpAdminService) getAdminId(cs *domain.CorpSigning) (string, error) {
	v, err := s.repo.FindCorpManagers(cs.Link.Id, cs.PrimaryEmailDomain())
	if err != nil {
		return "", err
	}

	m := map[string]bool{}
	for i := range v {
		m[v[i].Id] = true
	}

	r := domain.RoleAdmin

	for i := 0; m[r]; {
		i++
		r = domain.RoleAdmin + strconv.Itoa(i)
	}

	return r, nil
}
