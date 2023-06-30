package repository

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type CorpSigningSummary struct {
	Id     string
	Date   string
	HasPDF bool
	Link   domain.Link
	Rep    domain.Representative
	Corp   domain.Corporation
	Admin  domain.Manager
}

type EmployeeSigningSummary struct {
	Enabled bool
}

type CorpSigning interface {
	Add(*domain.CorpSigning) error
	// count the corp by the email domain
	Count(linkId, domain string) (int, error)
	Find(string) (domain.CorpSigning, error)
	Remove(*domain.CorpSigning) error
	FindAll(linkId string) ([]CorpSigningSummary, error)

	AddEmployee(*domain.CorpSigning, *domain.EmployeeSigning) error
	SaveEmployee(*domain.CorpSigning, *domain.EmployeeSigning) error
	FindEmployees(string) ([]domain.EmployeeSigning, error)
	RemoveEmployee(*domain.CorpSigning, *domain.EmployeeSigning) error
	FindEmployeesByEmail(linkId string, email dp.EmailAddr) ([]EmployeeSigningSummary, error)

	AddAdmin(*domain.CorpSigning) error

	AddEmployeeManagers(*domain.CorpSigning, []domain.Manager) error
	RemoveEmployeeManagers(*domain.CorpSigning, []string) error
	FindEmployeeManagers(string) ([]domain.Manager, error)

	AddEmailDomain(*domain.CorpSigning, string) error
	FindEmailDomains(string) ([]string, error)

	SaveCorpPDF(*domain.CorpSigning, []byte) error
	FindCorpPDF(string) ([]byte, error)
}
