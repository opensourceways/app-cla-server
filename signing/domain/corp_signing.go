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
	Id        string
	PDF       string
	Date      string
	Link      Link
	Rep       Representative
	Corp      Corporation
	Admin     Manager
	AllInfo   AllSingingInfo
	Employees []EmployeeSigning
	Version   int
}

func (cs *CorpSigning) PrimaryEmailDomain() string {
	return cs.Corp.PrimaryEmailDomain
}

func (cs *CorpSigning) HasAdmin() bool {
	return !cs.Admin.isEmpty()
}

func (cs *CorpSigning) SetAdmin(n int) error {
	if cs.HasAdmin() {
		return NewDomainError(ErrorCodeCorpAdminExists)
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

func (cs *CorpSigning) NewestEmployee() *EmployeeSigning {
	if n := len(cs.Employees); n > 0 {
		return &cs.Employees[n-1]
	}

	return nil
}
