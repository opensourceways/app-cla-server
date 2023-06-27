package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
)

type CmdToAddEmployeeManager struct {
	CorpSigningId string
	Managers      []domain.Manager
}

type CmdToRemoveEmployeeManager struct {
	CorpSigningId string
	Managers      []string
}

type RemovedManagerDTO struct {
	Name  string
	Email string
}

type EmployeeManagerDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func toEmployeeManagerDTO(m *domain.Manager) EmployeeManagerDTO {
	return EmployeeManagerDTO{
		ID:    m.Id,
		Name:  m.Name.Name(),
		Email: m.EmailAddr.EmailAddr(),
	}
}
