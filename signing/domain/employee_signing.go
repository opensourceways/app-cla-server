package domain

import (
	"github.com/opensourceways/app-cla-server/util"
)

const (
	employeeSigningActionEnable  = "enable"
	employeeSigningActionDisable = "disable"
	employeeSigningActionDelete  = "delete"
)

type EmployeeSigningLog struct {
	Time   int64
	Action string
}

type EmployeeSigning struct {
	Id      string
	CLA     CLA
	Rep     Representative
	Date    string
	Enabled bool
	Deleted bool
	AllInfo AllSingingInfo
	Logs    []EmployeeSigningLog
}

func (es *EmployeeSigning) isMe(es1 *EmployeeSigning) bool {
	return es.Rep.EmailAddr.EmailAddr() == es1.Rep.EmailAddr.EmailAddr()
}

func (es *EmployeeSigning) Enable() error {
	if es.Enabled {
		return NewDomainError(ErrorCodeEmployeeSigningEnableAgain)
	}

	es.Enabled = true
	es.addLog(employeeSigningActionEnable)

	return nil
}

func (es *EmployeeSigning) Disable() error {
	if !es.Enabled {
		return NewDomainError(ErrorCodeEmployeeSigningDisableAgain)
	}

	es.Enabled = false
	es.addLog(employeeSigningActionDisable)

	return nil
}

func (es *EmployeeSigning) Delete() error {
	if es.Enabled || es.Deleted {
		return NewDomainError(ErrorCodeEmployeeSigningCanNotDelete)
	}

	es.Deleted = true
	es.addLog(employeeSigningActionDelete)

	return nil
}

func (es *EmployeeSigning) addLog(action string) {
	es.Logs = append(es.Logs, EmployeeSigningLog{
		Time:   util.Now(),
		Action: action,
	})
}
