package repository

import "github.com/opensourceways/app-cla-server/signing/domain"

type CorpSigning interface {
	Add(*domain.CorpSigning) error
	AddEmployee(*domain.CorpSigning, *domain.EmployeeSigning) error
	AddAdmin(*domain.CorpSigning) error
	AddEmployeeManagers(*domain.CorpSigning, []domain.Manager) error

	// count the corp by the email domain
	Count(linkId, domain string) (int, error)
	Find(string) (domain.CorpSigning, error)
}
