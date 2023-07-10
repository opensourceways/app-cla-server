package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

func NewEmployeeSigningService(
	repo repository.CorpSigning,
	vc vcservice.VCService,
) *employeeSigningService {
	return &employeeSigningService{
		repo: repo,
		vc:   verificationCodeService{vc},
	}
}

type EmployeeSigningService interface {
	Verify(cmd *CmdToCreateVerificationCode) (string, error)
	Sign(cmd *CmdToSignEmployeeCLA) ([]EmployeeManagerDTO, error)
	Remove(cmd *CmdToRemoveEmployeeSigning) (string, error)
	Update(cmd *CmdToUpdateEmployeeSigning) (string, error)
	List(csId string) ([]EmployeeSigningDTO, error)
}

type employeeSigningService struct {
	vc   verificationCodeService
	repo repository.CorpSigning
}

func (s *employeeSigningService) Verify(cmd *CmdToCreateVerificationCode) (string, error) {
	return s.vc.newCode((*cmdToCreateCodeForEmployeeSigning)(cmd))
}

// Sign
func (s *employeeSigningService) Sign(cmd *CmdToSignEmployeeCLA) ([]EmployeeManagerDTO, error) {
	cmd1 := cmd.toCmd()
	if err := s.vc.validate(&cmd1, cmd.VerificationCode); err != nil {
		return nil, err
	}

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

	// TODO critical case that a employee signs two corps. A lock will fix it.
	v, err := s.repo.FindEmployeesByEmail(cs.Link.Id, cmd.Rep.EmailAddr)
	if err != nil {
		return nil, err
	}
	if len(v) > 0 {
		return nil, domain.NewNotFoundDomainError(domain.ErrorCodeEmployeeSigningReSigning)
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

// Remove
func (s *employeeSigningService) Remove(cmd *CmdToRemoveEmployeeSigning) (string, error) {
	cs, err := s.repo.Find(cmd.CorpSigningId)
	if err != nil {
		return "", err
	}

	es, err := cs.RemoveEmployee(cmd.EmployeeSigningId)
	if err != nil {
		return "", err
	}

	if err := s.repo.RemoveEmployee(&cs, es); err != nil {
		return "", err
	}

	return es.Rep.EmailAddr.EmailAddr(), nil
}

// Update
func (s *employeeSigningService) Update(cmd *CmdToUpdateEmployeeSigning) (string, error) {
	cs, err := s.repo.Find(cmd.CorpSigningId)
	if err != nil {
		return "", err
	}

	es, err := cs.UpdateEmployee(cmd.EmployeeSigningId, cmd.Enabled)
	if err != nil {
		return "", err
	}

	if err := s.repo.SaveEmployee(&cs, es); err != nil {
		return "", err
	}

	return es.Rep.EmailAddr.EmailAddr(), nil
}

// List
func (s *employeeSigningService) List(csId string) ([]EmployeeSigningDTO, error) {
	v, err := s.repo.FindEmployees(csId)
	if err != nil || len(v) == 0 {
		return nil, err
	}

	r := make([]EmployeeSigningDTO, len(v))
	for i := range v {
		r[i] = s.toEmployeeSigningDTO(&v[i])
	}

	return r, nil
}

func (s *employeeSigningService) toEmployeeSigningDTO(v *domain.EmployeeSigning) EmployeeSigningDTO {
	dto := IndividualSigningDTO{
		ID:    v.Id,
		Name:  v.Rep.Name.Name(),
		Date:  v.Date,
		Email: v.Rep.EmailAddr.EmailAddr(),
	}

	return EmployeeSigningDTO{
		IndividualSigningDTO: dto,
		Enabled:              v.Enabled,
	}
}
