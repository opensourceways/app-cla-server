package domain

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestNewCollectedEmployeeSigning(t *testing.T) {
	cla := CLAInfo{CLAId: "cla1", Language: dp.CreateLanguage("en")}
	rep := Representative{
		Name:      dp.CreateName("stu"),
		EmailAddr: dp.CreateEmailAddr("stu@uni.edu.cn"),
	}
	all := AllSingingInfo{"name": "stu"}

	es := NewCollectedEmployeeSigning(cla, rep, "2026-01-01", all)

	if es.Enabled {
		t.Error("collected employee signing must be disabled")
	}
	if es.Source != EmployeeSigningSourceIndividual {
		t.Errorf("Source = %q, want %q", es.Source, EmployeeSigningSourceIndividual)
	}
	if es.CLA != cla {
		t.Errorf("CLA = %v, want %v", es.CLA, cla)
	}
	if es.Rep.EmailAddr.EmailAddr() != "stu@uni.edu.cn" {
		t.Errorf("email = %s", es.Rep.EmailAddr.EmailAddr())
	}
	if es.Date != "2026-01-01" {
		t.Errorf("Date = %s, want the original individual signing date", es.Date)
	}
	if len(es.Logs) != 1 || es.Logs[0].Action != "collect" {
		t.Errorf("Logs = %v, want one collect entry", es.Logs)
	}
	if es.Logs[0].Time <= 0 {
		t.Errorf("collect log time = %d, want a unix timestamp", es.Logs[0].Time)
	}
}

func TestCorpSigningCollectEmployee(t *testing.T) {
	newCorp := func(domains ...string) *CorpSigning {
		return &CorpSigning{
			Corp: Corporation{
				AllEmailDomains:    domains,
				PrimaryEmailDomain: domains[0],
			},
		}
	}

	t.Run("success without managers", func(t *testing.T) {
		cs := newCorp("uni.edu.cn")

		es := NewCollectedEmployeeSigning(
			CLAInfo{CLAId: "cla1"},
			Representative{EmailAddr: dp.CreateEmailAddr("stu@uni.edu.cn")},
			"2026-01-01", nil,
		)

		if err := cs.CollectEmployee(&es); err != nil {
			t.Fatalf("CollectEmployee: unexpected error %v", err)
		}
		if len(cs.Employees) != 1 {
			t.Fatalf("employees = %d, want 1", len(cs.Employees))
		}
		if !cs.Employees[0].isMe(&es) {
			t.Error("the collected employee signing should be appended")
		}
	})

	t.Run("rejects email of other corp", func(t *testing.T) {
		cs := newCorp("uni.edu.cn")

		es := NewCollectedEmployeeSigning(
			CLAInfo{CLAId: "cla1"},
			Representative{EmailAddr: dp.CreateEmailAddr("stu@other.com")},
			"2026-01-01", nil,
		)

		if err := cs.CollectEmployee(&es); err == nil {
			t.Error("CollectEmployee should fail for an email outside the corp domains")
		}
		if len(cs.Employees) != 0 {
			t.Errorf("employees = %d, want 0", len(cs.Employees))
		}
	})

	t.Run("rejects duplicated employee", func(t *testing.T) {
		cs := newCorp("uni.edu.cn")
		cs.Employees = []EmployeeSigning{
			{Id: "e1", Rep: Representative{EmailAddr: dp.CreateEmailAddr("stu@uni.edu.cn")}},
		}

		es := NewCollectedEmployeeSigning(
			CLAInfo{CLAId: "cla1"},
			Representative{EmailAddr: dp.CreateEmailAddr("stu@uni.edu.cn")},
			"2026-01-01", nil,
		)

		if err := cs.CollectEmployee(&es); err == nil {
			t.Error("CollectEmployee should fail for a duplicated email")
		}
		if len(cs.Employees) != 1 {
			t.Errorf("employees = %d, want 1", len(cs.Employees))
		}
	})

	t.Run("domain matching is case insensitive", func(t *testing.T) {
		cs := newCorp("UNI.edu.cn")

		es := NewCollectedEmployeeSigning(
			CLAInfo{CLAId: "cla1"},
			Representative{EmailAddr: dp.CreateEmailAddr("stu@uni.edu.cn")},
			"2026-01-01", nil,
		)

		if err := cs.CollectEmployee(&es); err != nil {
			t.Fatalf("CollectEmployee: unexpected error %v", err)
		}
	})
}

func TestCorporationIsMyEmailIgnoreCase(t *testing.T) {
	corp := Corporation{AllEmailDomains: []string{"UNI.edu.cn", "Other.com"}}

	if !corp.isMyEmailIgnoreCase(dp.CreateEmailAddr("user@uni.edu.cn")) {
		t.Error("should match ignoring case")
	}
	if !corp.isMyEmailIgnoreCase(dp.CreateEmailAddr("user@OTHER.com")) {
		t.Error("should match ignoring case")
	}
	if corp.isMyEmailIgnoreCase(dp.CreateEmailAddr("user@unknown.com")) {
		t.Error("should not match unknown domain")
	}
}
