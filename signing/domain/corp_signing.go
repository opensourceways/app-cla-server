package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
)

type AllSingingInfo = map[string]string

type Representative struct {
	Name      dp.Name
	EmailAddr dp.EmailAddr
}

type CLAInfo struct {
	CLAId    string
	Language dp.Language
}

type LinkInfo struct {
	Id string

	CLAInfo
}

type CorpSigning struct {
	Id      string
	Date    string
	Link    LinkInfo
	Rep     Representative
	Corp    Corporation
	AllInfo AllSingingInfo

	HasPDF    bool // true if pdf has uploaded
	Admin     Manager
	Managers  []Manager
	Employees []EmployeeSigning
	Version   int
}

func (cs *CorpSigning) CorpName() dp.CorpName {
	return cs.Corp.Name
}

func (cs *CorpSigning) GetRole(e dp.EmailAddr) string {
	if cs.Admin.hasEmail(e) {
		return RoleAdmin
	}

	for i := range cs.Managers {
		if cs.Managers[i].hasEmail(e) {
			return RoleManager
		}
	}

	return ""
}

func (cs *CorpSigning) PrimaryEmailDomain() string {
	return cs.Corp.PrimaryEmailDomain
}

func (cs *CorpSigning) CanRemove() error {
	if !cs.Admin.IsEmpty() {
		return NewDomainError(ErrorCodeCorpSigningCanNotDelete)
	}

	return nil
}

func (cs *CorpSigning) CanSetAdmin() error {
	if !cs.HasPDF {
		return NewNotFoundDomainError(ErrorCodeCorpPDFNotFound)
	}

	if !cs.Admin.IsEmpty() {
		return NewDomainError(ErrorCodeCorpAdminExists)
	}

	return nil
}

func (cs *CorpSigning) SetAdmin(adminId string) {
	cs.Admin.Id = adminId
	cs.Admin.Representative = cs.Rep
}

func (cs *CorpSigning) AddEmailDomain(email dp.EmailAddr) error {
	return cs.Corp.addEmailDomain(email.Domain())
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

func (cs *CorpSigning) RemoveManagers(managers []string) ([]Manager, error) {
	if len(managers) > config.MaxNumOfEmployeeManager {
		return nil, NewDomainError(ErrorCodeEmployeeManagerTooMany)
	}

	toRemove := make(map[int]bool)

	for i := range managers {
		j, exists := cs.posOfManager(managers[i])
		if !exists {
			return nil, NewDomainError(ErrorCodeEmployeeManagerNotExists)
		}
		toRemove[j] = true
	}

	var r = []Manager{}

	if n := len(cs.Managers) - len(toRemove); n <= 0 {
		r = cs.Managers
		cs.Managers = nil
	} else {
		m := make([]Manager, 0, n)
		r = make([]Manager, 0, len(toRemove))

		for i := range cs.Managers {
			if toRemove[i] {
				r = append(r, cs.Managers[i])
			} else {
				m = append(m, cs.Managers[i])
			}
		}

		cs.Managers = m
	}

	return r, nil
}

func (cs *CorpSigning) AddEmployee(es *EmployeeSigning) error {
	if len(cs.Managers) == 0 {
		return NewDomainError(ErrorCodeEmployeeSigningNoManager)
	}

	if !cs.isSameCorp(es.Rep.EmailAddr) {
		return NewDomainError(ErrorCodeEmployeeNotSameCorp)
	}

	for i := range cs.Employees {
		if cs.Employees[i].isMe(es) {
			return NewDomainError(ErrorCodeEmployeeSigningReSigning)
		}
	}

	cs.Employees = append(cs.Employees, *es)

	return nil
}

func (cs *CorpSigning) UpdateEmployee(index string, enabled bool) (es *EmployeeSigning, err error) {
	i, ok := cs.posOfEmployee(index)
	if !ok {
		err = NewDomainError(ErrorCodeEmployeeSigningNotFound)

		return
	}

	es = &cs.Employees[i]

	if enabled {
		err = es.enable()
	} else {
		err = es.disable()
	}

	return
}

func (cs *CorpSigning) RemoveEmployee(index string) (es *EmployeeSigning, err error) {
	i, ok := cs.posOfEmployee(index)
	if !ok {
		err = NewDomainError(ErrorCodeEmployeeSigningNotFound)

		return
	}

	es = &cs.Employees[i]

	err = es.remove()

	return
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

func (cs *CorpSigning) posOfManager(index string) (int, bool) {
	for j := range cs.Managers {
		if cs.Managers[j].Id == index {
			return j, true
		}
	}

	return 0, false
}

func (cs *CorpSigning) posOfEmployee(index string) (int, bool) {
	for j := range cs.Employees {
		if cs.Employees[j].Id == index {
			return j, true
		}
	}

	return 0, false
}
