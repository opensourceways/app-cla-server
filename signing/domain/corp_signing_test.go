package domain

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestCorpSigningHasSignedCLA(t *testing.T) {
	cs := &CorpSigning{Link: LinkInfo{CLAInfo: CLAInfo{CLAId: "10"}}}

	if !cs.HasSignedCLA("10") {
		t.Error("HasSignedCLA should return true when CLAId matches")
	}
	if cs.HasSignedCLA("11") {
		t.Error("HasSignedCLA should return false when CLAId does not match")
	}
}

func TestCorpSigningSetLatestClaIdWithExistingLogs(t *testing.T) {
	cs := &CorpSigning{
		Link:           LinkInfo{CLAInfo: CLAInfo{CLAId: "3"}},
		PendingCLAId:   "5",
		ClaNotifyCount: 3,
		ClaNotifyTime:  2000,
	}

	err := cs.SetLatestClaId("5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cs.Link.CLAId != "5" {
		t.Errorf("CLAId: got %s, want 5", cs.Link.CLAId)
	}
	if cs.PendingCLAId != "" {
		t.Errorf("PendingCLAId: got %s, want empty", cs.PendingCLAId)
	}
	if cs.ClaNotifyCount != 0 {
		t.Errorf("ClaNotifyCount: got %d, want 0", cs.ClaNotifyCount)
	}
	if cs.ClaNotifyTime != 0 {
		t.Errorf("ClaNotifyTime: got %d, want 0", cs.ClaNotifyTime)
	}
}

func TestCorpSigningCanRemove(t *testing.T) {
	cs1 := &CorpSigning{Admin: Manager{Id: "admin1"}}
	if cs1.CanRemove() == nil {
		t.Error("CanRemove should fail when admin exists")
	}

	cs2 := &CorpSigning{Admin: Manager{}}
	if cs2.CanRemove() != nil {
		t.Error("CanRemove should succeed when admin is empty")
	}
}

func TestCorpSigningCanSetAdmin(t *testing.T) {
	tests := []struct {
		name    string
		hasPDF  bool
		adminId string
		wantErr bool
	}{
		{"no PDF", false, "", true},
		{"admin exists", true, "admin1", true},
		{"can set", true, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &CorpSigning{HasPDF: tt.hasPDF, Admin: Manager{Id: tt.adminId}}
			err := cs.CanSetAdmin()
			if (err != nil) != tt.wantErr {
				t.Errorf("CanSetAdmin() err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestCorpSigningSetAdmin(t *testing.T) {
	cs := &CorpSigning{
		Rep: Representative{
			Name:      dp.CreateName("John"),
			EmailAddr: dp.CreateEmailAddr("john@example.com"),
		},
	}

	cs.SetAdmin("admin123")

	if cs.Admin.Id != "admin123" {
		t.Errorf("Admin.Id: got %s, want admin123", cs.Admin.Id)
	}
	if cs.Admin.Representative.Name.Name() != "John" {
		t.Errorf("Admin.Name: got %s, want John", cs.Admin.Name.Name())
	}
}

func TestCorpSigningIsAdmin(t *testing.T) {
	email := dp.CreateEmailAddr("admin@test.com")
	cs := &CorpSigning{Admin: Manager{Id: "1", Representative: Representative{EmailAddr: email}}}

	if !cs.IsAdmin(email) {
		t.Error("IsAdmin should return true for admin email")
	}

	other := dp.CreateEmailAddr("other@test.com")
	if cs.IsAdmin(other) {
		t.Error("IsAdmin should return false for non-admin email")
	}
}

func TestCorpSigningGetRole(t *testing.T) {
	adminEmail := dp.CreateEmailAddr("admin@test.com")
	managerEmail := dp.CreateEmailAddr("mgr@test.com")

	cs := &CorpSigning{
		Admin:    Manager{Id: "1", Representative: Representative{EmailAddr: adminEmail}},
		Managers: []Manager{{Id: "2", Representative: Representative{EmailAddr: managerEmail}}},
	}

	if cs.GetRole(adminEmail) != RoleAdmin {
		t.Errorf("GetRole admin: got %s, want %s", cs.GetRole(adminEmail), RoleAdmin)
	}
	if cs.GetRole(managerEmail) != RoleManager {
		t.Errorf("GetRole manager: got %s, want %s", cs.GetRole(managerEmail), RoleManager)
	}

	other := dp.CreateEmailAddr("unknown@test.com")
	if cs.GetRole(other) != "" {
		t.Errorf("GetRole unknown: got %s, want empty", cs.GetRole(other))
	}
}

func TestCorpSigningPrimaryEmailDomain(t *testing.T) {
	cs := &CorpSigning{Corp: Corporation{PrimaryEmailDomain: "example.com"}}
	if cs.PrimaryEmailDomain() != "example.com" {
		t.Errorf("PrimaryEmailDomain: got %s, want example.com", cs.PrimaryEmailDomain())
	}
}

func TestCorpSigningCorpName(t *testing.T) {
	cs := &CorpSigning{Corp: Corporation{Name: dp.CreateCorpName("TestCorp")}}
	if cs.CorpName().CorpName() != "TestCorp" {
		t.Errorf("CorpName: got %s, want TestCorp", cs.CorpName().CorpName())
	}
}

func TestCorpSigningAddEmployee(t *testing.T) {
	cs := &CorpSigning{
		Corp:     Corporation{AllEmailDomains: []string{"test.com"}},
		Managers: []Manager{{Id: "m1"}},
	}

	es := &EmployeeSigning{
		Rep: Representative{EmailAddr: dp.CreateEmailAddr("user@test.com")},
	}

	if err := cs.AddEmployee(es); err != nil {
		t.Fatalf("AddEmployee first: unexpected error %v", err)
	}
	if len(cs.Employees) != 1 {
		t.Errorf("Employees length: got %d, want 1", len(cs.Employees))
	}

	if err := cs.AddEmployee(es); err == nil {
		t.Error("AddEmployee duplicate should fail")
	}
}

func TestCorpSigningAddEmployeeNoManager(t *testing.T) {
	cs := &CorpSigning{
		Corp:     Corporation{AllEmailDomains: []string{"test.com"}},
		Managers: []Manager{},
	}

	es := &EmployeeSigning{
		Rep: Representative{EmailAddr: dp.CreateEmailAddr("user@test.com")},
	}

	if err := cs.AddEmployee(es); err == nil {
		t.Error("AddEmployee with no manager should fail")
	}
}

func TestCorpSigningAddEmployeeNotSameCorp(t *testing.T) {
	cs := &CorpSigning{
		Corp:     Corporation{AllEmailDomains: []string{"corp.com"}},
		Managers: []Manager{{Id: "m1"}},
	}

	es := &EmployeeSigning{
		Rep: Representative{EmailAddr: dp.CreateEmailAddr("user@other.com")},
	}

	if err := cs.AddEmployee(es); err == nil {
		t.Error("AddEmployee with different corp should fail")
	}
}

func TestIsErrorOf(t *testing.T) {
	err := NewDomainError(ErrorCodeCorpSigningCLAIsLatest)
	if !IsErrorOf(err, ErrorCodeCorpSigningCLAIsLatest) {
		t.Error("IsErrorOf should match error code")
	}
	if IsErrorOf(err, "other_code") {
		t.Error("IsErrorOf should not match different code")
	}
}

func TestDomainErrorError(t *testing.T) {
	err := domainError("test_error_code")
	if err.Error() != "test error code" {
		t.Errorf("domainError.Error(): got %s, want 'test error code'", err.Error())
	}
}

func TestDomainErrorErrorCode(t *testing.T) {
	err := domainError("test_error_code")
	if err.ErrorCode() != "test_error_code" {
		t.Errorf("domainError.ErrorCode(): got %s, want 'test_error_code'", err.ErrorCode())
	}
}

func TestNotFoundDomainError(t *testing.T) {
	err := NewNotFoundDomainError(ErrorCodeCorpPDFNotFound)
	err.NotFound()

	if !IsErrorOf(err, ErrorCodeCorpPDFNotFound) {
		t.Error("notFoundError should still match error code via IsErrorOf")
	}
}
