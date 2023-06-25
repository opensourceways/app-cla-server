package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

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
	AllInfo   AllSingingInfo
	Employees []EmployeeSigning
	Version   int
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
