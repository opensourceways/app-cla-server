package domain

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestEmployeeSigningIsMe(t *testing.T) {
	email := dp.CreateEmailAddr("user@test.com")
	es := &EmployeeSigning{Rep: Representative{EmailAddr: email}}

	es2 := &EmployeeSigning{Rep: Representative{EmailAddr: email}}
	if !es.isMe(es2) {
		t.Error("isMe should return true for same email")
	}

	es3 := &EmployeeSigning{Rep: Representative{EmailAddr: dp.CreateEmailAddr("other@test.com")}}
	if es.isMe(es3) {
		t.Error("isMe should return false for different email")
	}
}

func TestEmployeeSigningEnable(t *testing.T) {
	es := &EmployeeSigning{Enabled: false, Logs: []EmployeeSigningLog{}}

	if err := es.enable(); err != nil {
		t.Fatalf("enable: unexpected error %v", err)
	}
	if !es.Enabled {
		t.Error("Enabled should be true after enable")
	}
	if len(es.Logs) != 1 || es.Logs[0].Action != "enable" {
		t.Errorf("Logs after enable: got %v", es.Logs)
	}

	if err := es.enable(); err == nil {
		t.Error("enable again should fail")
	}
}

func TestEmployeeSigningDisable(t *testing.T) {
	es := &EmployeeSigning{Enabled: true, Logs: []EmployeeSigningLog{}}

	if err := es.disable(); err != nil {
		t.Fatalf("disable: unexpected error %v", err)
	}
	if es.Enabled {
		t.Error("Enabled should be false after disable")
	}
	if len(es.Logs) != 1 || es.Logs[0].Action != "disable" {
		t.Errorf("Logs after disable: got %v", es.Logs)
	}

	if err := es.disable(); err == nil {
		t.Error("disable again should fail")
	}
}

func TestEmployeeSigningRemove(t *testing.T) {
	es1 := &EmployeeSigning{Enabled: false}
	if err := es1.remove(); err != nil {
		t.Fatalf("remove disabled: unexpected error %v", err)
	}

	es2 := &EmployeeSigning{Enabled: true}
	if err := es2.remove(); err == nil {
		t.Error("remove enabled should fail")
	}
}

func TestCorpSigningUpdateEmployee(t *testing.T) {
	email := dp.CreateEmailAddr("user@test.com")
	cs := &CorpSigning{
		Employees: []EmployeeSigning{
			{Id: "emp1", Enabled: false, Rep: Representative{EmailAddr: email}},
		},
	}

	es, err := cs.UpdateEmployee("emp1", true)
	if err != nil {
		t.Fatalf("UpdateEmployee enable: unexpected error %v", err)
	}
	if !es.Enabled {
		t.Error("Employee should be enabled")
	}

	_, err = cs.UpdateEmployee("unknown", true)
	if err == nil {
		t.Error("UpdateEmployee unknown id should fail")
	}
}

func TestCorpSigningRemoveEmployee(t *testing.T) {
	cs := &CorpSigning{
		Employees: []EmployeeSigning{
			{Id: "emp1", Enabled: false},
		},
	}

	es, err := cs.RemoveEmployee("emp1")
	if err != nil {
		t.Fatalf("RemoveEmployee: unexpected error %v", err)
	}
	if es == nil {
		t.Error("RemoveEmployee should return employee")
	}

	_, err = cs.RemoveEmployee("unknown")
	if err == nil {
		t.Error("RemoveEmployee unknown id should fail")
	}
}
