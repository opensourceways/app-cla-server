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
	Sign(cmd *CmdToSignEmployeeCLA) error
}

type employeeSigningService struct {
	repo repository.CorpSigning
}

func (s *employeeSigningService) Sign(cmd *CmdToSignEmployeeCLA) error {
	cs, err := s.repo.Find(cmd.CorpSigningId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = domain.NewNotFoundDomainError(domain.ErrorCodeCorpSigningNotFound)
		}

		return err
	}

	es := cmd.toEmployeeSigning()
	if err := cs.AddEmployee(&es); err != nil {
		return err
	}

	return s.repo.AddEmployee(&cs)
}
