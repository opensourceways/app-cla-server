package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
)

func NewEmployeeManagerService(
	repo repository.CorpSigning,
	userService userservice.UserService,
) EmployeeManagerService {
	return &employeeManagerService{
		repo:        repo,
		userService: userService,
	}
}

type EmployeeManagerService interface {
	Add(cmd *CmdToAddEmployeeManager) ([]ManagerDTO, error)
	Remove(cmd *CmdToRemoveEmployeeManager) (dtos []RemovedManagerDTO, err error)
}

type employeeManagerService struct {
	repo        repository.CorpSigning
	userService userservice.UserService
}

func (s *employeeManagerService) Add(cmd *CmdToAddEmployeeManager) ([]ManagerDTO, error) {
	cs, err := s.repo.Find(cmd.CorpSigningId)
	if err != nil {
		return nil, err
	}

	if err = cs.AddManagers(cmd.Managers); err != nil {
		return nil, err
	}

	pws, ids, err := s.userService.Add(cs.Link.Id, cmd.CorpSigningId, cmd.Managers)
	if err != nil {
		return nil, err
	}

	if err = s.repo.AddEmployeeManagers(&cs, cmd.Managers); err != nil {
		s.userService.Remove(ids)
	}

	return s.toManageDTOs(pws, cmd.Managers)
}

func (s *employeeManagerService) Remove(cmd *CmdToRemoveEmployeeManager) (dtos []RemovedManagerDTO, err error) {
	cs, err := s.repo.Find(cmd.CorpSigningId)
	if err != nil {
		return
	}

	removed, err := cs.RemoveManagers(cmd.Managers)
	if err != nil {
		return
	}

	if err = s.repo.RemoveEmployeeManagers(&cs, cmd.Managers); err != nil {
		return
	}

	accounts := make([]dp.Account, len(removed))
	dtos = make([]RemovedManagerDTO, len(removed))

	for i := range removed {
		item := &removed[i]

		if accounts[i], err = item.Account(); err != nil {
			return
		}

		dtos[i] = RemovedManagerDTO{
			Name:  item.Name.Name(),
			Email: item.EmailAddr.EmailAddr(),
		}
	}

	s.userService.RemoveByAccount(cs.Link.Id, accounts)

	return
}

func (s *employeeManagerService) toManageDTOs(pws map[string]string, ms []domain.Manager) ([]ManagerDTO, error) {
	dtos := make([]ManagerDTO, len(ms))

	for i := range ms {
		item := &ms[i]

		account, err := item.Account()
		if err != nil {
			return nil, err
		}

		dtos[i] = ManagerDTO{
			Role:      domain.RoleManager,
			Name:      item.Name.Name(),
			Account:   account.Account(),
			Password:  pws[item.Id],
			EmailAddr: item.EmailAddr.EmailAddr(),
		}
	}

	return dtos, nil
}
