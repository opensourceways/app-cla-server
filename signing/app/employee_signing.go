package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewEmployeeSigningService(repo repository.CorpSigning) *employeeSigningService {
	return &employeeSigningService{repo}
}

type EmployeeSigningService interface {
	Sign(cmd *CmdToSignEmployeeCLA) ([]EmployeeManagerDTO, error)
	Update(cmd CmdToUpdateEmployeeSigning) ([]EmployeeManagerDTO, error)
}

type employeeSigningService struct {
	repo repository.CorpSigning
}

func (s *employeeSigningService) Sign(cmd *CmdToSignEmployeeCLA) ([]EmployeeManagerDTO, error) {
	cs, err := s.repo.Find(cmd.CorpSigningId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = domain.NewNotFoundDomainError(domain.ErrorCodeCorpSigningNotFound)
		}

		return nil, err
	}

	es := cmd.toEmployeeSigning()
	if err := cs.AddEmployee(&es); err != nil {
		return nil, err
	}

	if err := s.repo.AddEmployee(&cs, &es); err != nil {
		return nil, err
	}

	dtos := make([]EmployeeManagerDTO, len(cs.Managers))
	for i := range cs.Managers {
		dtos[i] = toEmployeeManagerDTO(&cs.Managers[i])
	}

	return dtos, nil
}

func (s *employeeSigningService) Update(cmd CmdToUpdateEmployeeSigning) ([]EmployeeManagerDTO, error) {
	cs, err := s.repo.Find(cmd.CorpSigningId)
	if err != nil {
		return nil, err
	}

	es, err := cs.UpdateEmployee(cmd.EmployeeSigningId, cmd.Enabled)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveEmployee(&cs, es); err != nil {
		return nil, err
	}

	dtos := make([]EmployeeManagerDTO, len(cs.Managers))
	for i := range cs.Managers {
		dtos[i] = toEmployeeManagerDTO(&cs.Managers[i])
	}

	return dtos, nil
}
