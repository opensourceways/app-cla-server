package domain

import (
	"strconv"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
)

type AllSingingInfo = map[string]string

type Representative struct {
	Name      dp.Name
	EmailAddr dp.EmailAddr
}

type CLA struct {
	CLAId    string
	Language dp.Language
}

type Link struct {
	Id string

	CLA
}

type CorpSigning struct {
	Id      string
	PDF     string
	Date    string
	Link    Link
	Rep     Representative
	Corp    Corporation
	AllInfo AllSingingInfo

	Admin     Manager
	Managers  []Manager
	Employees []EmployeeSigning
	Version   int
}

func (cs *CorpSigning) PrimaryEmailDomain() string {
	return cs.Corp.PrimaryEmailDomain
}

func (cs *CorpSigning) CanSetAdmin() error {
	if cs.PDF == "" {
		return NewNotFoundDomainError(ErrorCodeCorpPDFNotFound)
	}

	if !cs.Admin.isEmpty() {
		return NewDomainError(ErrorCodeCorpAdminExists)
	}

	return nil
}

func (cs *CorpSigning) SetAdmin(n int) error {
	if err := cs.CanSetAdmin(); err != nil {
		return err
	}

	v := RoleAdmin
	if n > 0 {
		v += strconv.Itoa(n)
	}

	cs.Admin.Id = v
	cs.Admin.Representative = cs.Rep

	return nil
}

func (cs *CorpSigning) AllEmailDomains() []string {
	return cs.Corp.AllEmailDomains
}

func (cs *CorpSigning) AddManagers(managers []Manager) error {
	if len(cs.Managers)+len(managers) > config.MaxNumOfEmployeeManager {
		return NewDomainError(ErrorCodeEmployeeManagerTooMany)
	}

	for i := range managers {
		item := &managers[i]

		if !cs.isSameCorp(item.EmailAddr) {
			return NewDomainError(ErrorCodeEmployeeManagerNotSameCorp)
		}

		if cs.hasManager(item) {
			return NewDomainError(ErrorCodeEmployeeManagerExists)
		}

		if cs.Admin.IsSame(item) {
			return NewDomainError(ErrorCodeEmployeeManagerAdminAsManager)
		}
	}

	return nil
}

func (cs *CorpSigning) AddEmployee(es *EmployeeSigning) error {
	// TODO manager

	for i := range cs.Employees {
		if cs.Employees[i].isMe(es) {
			return NewDomainError(ErrorCodeEmployeeSigningReSigning)
		}
	}

	cs.Employees = append(cs.Employees, *es)

	return nil
}

func (cs *CorpSigning) isSameCorp(email dp.EmailAddr) bool {
	return cs.Corp.isMyEmail(email)
}

func (cs *CorpSigning) hasManager(m *Manager) bool {
	for j := range cs.Managers {
		if cs.Managers[j].IsSame(m) {
			return true
		}
	}

	return false
}
