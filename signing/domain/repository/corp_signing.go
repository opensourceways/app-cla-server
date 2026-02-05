package repository

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type CorpSigningSummary struct {
	Id        string
	Date      string
	HasPDF    bool
	CLANotify string
	Link      domain.LinkInfo
	Rep       domain.Representative
	Corp      domain.Corporation
	Admin     domain.Manager
}

type EmployeeSigningSummary struct {
	Enabled bool
	ClaId   string
}

type CorpSigningSummaryPage struct {
	Total int64
	Data  []CorpSigningSummary
}

type CorpSummary struct {
	CorpName      dp.CorpName
	HasManager    bool
	CorpSigningId string
}

type CorpSigning interface {
	Update(cs *domain.CorpSigning) error
	Add(*domain.CorpSigning) error
	AddForMigrate(*domain.CorpSigning) error
	FindCorpSummary(linkId, domain string) ([]CorpSummary, error)
	FindCorpManagers(linkId, domain string) ([]domain.Manager, error)
	Find(string) (domain.CorpSigning, error)
	Remove(*domain.CorpSigning) error
	FindAll(linkId string) ([]CorpSigningSummary, error)
	FindPage(linkId string, intPage, intPageSize int, adminAdded bool, searchQuery string) (CorpSigningSummaryPage, error)
	FindAllWithPagination(linkId string, offset, limit int) ([]CorpSigningSummary, error)
	CountByLinkId(linkId string) (int64, error)

	AddEmployee(*domain.CorpSigning, *domain.EmployeeSigning) error
	SaveEmployee(*domain.CorpSigning, *domain.EmployeeSigning) error
	FindEmployees(string) ([]domain.EmployeeSigning, error)
	RemoveEmployee(*domain.CorpSigning, *domain.EmployeeSigning) error
	FindEmployeesByEmail(linkId string, email dp.EmailAddr) (EmployeeSigningSummary, error)

	AddAdmin(*domain.CorpSigning) error

	AddEmployeeManagers(*domain.CorpSigning, []domain.Manager) error
	RemoveEmployeeManagers(*domain.CorpSigning, []string) error
	FindEmployeeManagers(string) ([]domain.Manager, error)

	AddEmailDomain(*domain.CorpSigning, string) error
	FindEmailDomains(string) ([]string, error)

	SaveCorpPDF(*domain.CorpSigning, []byte) error
	FindCorpPDF(string) ([]byte, error)

	HasSignedLink(linkId string) (bool, error)
	HasSignedCLA(*domain.CLAIndex, dp.CLAType) (bool, error)
	UpdateClaId(cs *domain.CorpSigning) error
	UpdateCLANotify(summary *CorpSigningSummary) error
}
