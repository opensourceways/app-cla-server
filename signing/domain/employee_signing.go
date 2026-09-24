package domain

import (
	"github.com/opensourceways/app-cla-server/util"
)

const (
	employeeSigningActionEnable     = "enable"
	employeeSigningActionDisable    = "disable"
	employeeSigningActionDelete     = "delete"
	employeeSigningActionCollect    = "collect"
	employeeSigningActionAutoEnable = "auto_enable"
)

const (
	// EmployeeSigningSourceIndividual marks the employee signing is collected
	// from a historical individual signing.
	EmployeeSigningSourceIndividual = "individual"

	// EmployeeSigningSourceAutoEnable marks the employee singing is enable automatically.
	EmployeeSigningSourceAutoEnable = "auto_enable"
)

type EmployeeSigningLog struct {
	Time   int64
	Action string
}

type EmployeeSigning struct {
	Id      string
	CLA     CLAInfo
	Rep     Representative
	Date    string
	Enabled bool
	AllInfo AllSingingInfo
	Logs    []EmployeeSigningLog
	Source  string
}

// NewCollectedEmployeeSigning generates a disabled employee signing from a
// historical individual signing. It must be activated by a corp manager
// before it takes effect.
func NewCollectedEmployeeSigning(
	cla CLAInfo, rep Representative, date string, all AllSingingInfo,
) EmployeeSigning {
	es := EmployeeSigning{
		CLA:     cla,
		Rep:     rep,
		Date:    date,
		Enabled: false,
		AllInfo: all,
		Source:  EmployeeSigningSourceIndividual,
	}

	es.addLog(employeeSigningActionCollect)

	return es
}

func (es *EmployeeSigning) isMe(es1 *EmployeeSigning) bool {
	return es.Rep.EmailAddr.EmailAddr() == es1.Rep.EmailAddr.EmailAddr()
}

func (es *EmployeeSigning) enable() error {
	if es.Enabled {
		return NewDomainError(ErrorCodeEmployeeSigningEnableAgain)
	}

	es.Enabled = true
	es.addLog(employeeSigningActionEnable)

	return nil
}

// AutoEnable marks the employee signing as enabled by the auto-approval
// preference (corp admin opened the switch). Unlike enable(), it writes a
// distinct "auto_enable" audit log so auto-passed records can be told apart
// from manually approved ones. It is called on freshly created employee
// signings during Sign, so Enabled is always false at this point.
func (es *EmployeeSigning) AutoEnable() {
	es.Enabled = true
	es.Source = EmployeeSigningSourceAutoEnable
	es.addLog(employeeSigningActionAutoEnable)
}

func (es *EmployeeSigning) disable() error {
	if !es.Enabled {
		return NewDomainError(ErrorCodeEmployeeSigningDisableAgain)
	}

	es.Enabled = false
	es.addLog(employeeSigningActionDisable)

	return nil
}

func (es *EmployeeSigning) remove() error {
	if es.Enabled {
		return NewDomainError(ErrorCodeEmployeeSigningCanNotDelete)
	}

	es.addLog(employeeSigningActionDelete)

	return nil
}

func (es *EmployeeSigning) addLog(action string) {
	es.Logs = append(es.Logs, EmployeeSigningLog{
		Time:   util.Now(),
		Action: action,
	})
}
