package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
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

	return s.toManageDTOs(pws, cmd.Managers), nil
}

func (s *employeeManagerService) toManageDTOs(pws map[string]string, ms []domain.Manager) []ManagerDTO {
	dtos := make([]ManagerDTO, len(ms))

	for i := range ms {
		item := &ms[i]

		dtos[i] = ManagerDTO{
			Id:        item.Id,
			Role:      domain.RoleManager,
			Name:      item.Name.Name(),
			Password:  pws[item.Id],
			EmailAddr: item.EmailAddr.EmailAddr(),
		}
	}

	return dtos
}
