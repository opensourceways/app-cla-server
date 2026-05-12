package repositoryimpl

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestToLinkDOAndBack(t *testing.T) {
	link := &domain.Link{
		Id: "link-1",
		Org: domain.OrgInfo{
			Alias:      "TestOrg",
			Logo:       "logo.png",
			ProjectURL: "https://example.com",
		},
		Email: domain.EmailInfo{
			Addr:     dp.CreateEmailAddr("org@test.com"),
			Platform: "gmail",
		},
		Submitter: "admin",
		CLANum:    1,
		CLAs: []domain.CLA{
			{
				Id:       "0",
				URL:      dp.CreateURL("https://example.com/cla.pdf"),
				Type:     dp.CLATypeCorp,
				Language: dp.CreateLanguage("en"),
				Fields: []domain.Field{
					{Id: "f0", Required: true, CLAField: dp.CLAField{Type: "text", Title: "Name", Desc: "desc"}},
				},
			},
		},
	}

	do := toLinkDO(link)
	if do.Id != "link-1" {
		t.Errorf("expected 'link-1', got '%s'", do.Id)
	}
	if do.Org.Alias != "TestOrg" {
		t.Errorf("expected 'TestOrg', got '%s'", do.Org.Alias)
	}
	if do.Email.Addr != "org@test.com" {
		t.Errorf("expected 'org@test.com', got '%s'", do.Email.Addr)
	}

	result := do.toLink()
	if result.Id != "link-1" || result.Org.Alias != "TestOrg" {
		t.Errorf("roundtrip failed")
	}
}

func TestToOrgInfoDOAndBack(t *testing.T) {
	org := &domain.OrgInfo{Alias: "MyOrg", Logo: "logo.png", ProjectURL: "https://x.com"}
	do := toOrgInfoDO(org)
	result := do.toOrgInfo()
	if result.Alias != "MyOrg" {
		t.Errorf("expected 'MyOrg', got '%s'", result.Alias)
	}
}

func TestToEmailInfoDOAndBack(t *testing.T) {
	info := &domain.EmailInfo{Addr: dp.CreateEmailAddr("test@example.com"), Platform: "smtp"}
	do := toEmailInfoDO(info)
	if do.Platform != "smtp" {
		t.Errorf("expected 'smtp', got '%s'", do.Platform)
	}
	result := do.toEmailInfo()
	if result.Addr.EmailAddr() != "test@example.com" {
		t.Errorf("expected 'test@example.com', got '%s'", result.Addr.EmailAddr())
	}
}

func TestToFieldDOAndBack(t *testing.T) {
	f := &domain.Field{Id: "f1", Required: true, CLAField: dp.CLAField{Type: "text", Title: "Name", Desc: "Desc"}}
	do := toFieldDO(f)
	if do.Id != "f1" || do.Type != "text" {
		t.Errorf("toFieldDO mismatch")
	}
	result := do.toField()
	if result.Id != "f1" || !result.Required {
		t.Errorf("toField mismatch: %+v", result)
	}
}

func TestToCLADOAndBack(t *testing.T) {
	cla := &domain.CLA{
		Id:       "0",
		URL:      dp.CreateURL("https://x.com/a.pdf"),
		Type:     dp.CLATypeCorp,
		Language: dp.CreateLanguage("en"),
		Fields: []domain.Field{
			{Id: "f1", Required: true, CLAField: dp.CLAField{Type: "text", Title: "Name"}},
		},
	}
	do := toCLADO(cla)
	if do.Id != "0" || do.Type != "corporation" || do.URL != "https://x.com/a.pdf" {
		t.Errorf("toCLADO mismatch")
	}
	result := do.toCLA()
	if result.Id != "0" || result.Type.CLAType() != "corporation" {
		t.Errorf("toCLA mismatch")
	}
}

func TestClaContentDO(t *testing.T) {
	do := claContentDO{
		LinkId: "L1",
		CLAId:  "C1",
		Text:   []byte("content"),
	}
	if do.LinkId != "L1" || do.CLAId != "C1" {
		t.Errorf("claContentDO mismatch")
	}
}

func TestToRepDOAndBack(t *testing.T) {
	rep := &domain.Representative{
		Name:      dp.CreateName("John"),
		EmailAddr: dp.CreateEmailAddr("john@test.com"),
	}
	do := toRepDO(rep)
	if do.Name != "John" || do.Email != "john@test.com" {
		t.Errorf("toRepDO mismatch")
	}
	result := do.toRep()
	if result.Name.Name() != "John" {
		t.Errorf("toRep mismatch")
	}
}

func TestToRepDONil(t *testing.T) {
	do := toRepDO(nil)
	if do.Name != "" || do.Email != "" {
		t.Errorf("expected empty RepDO for nil")
	}

	var nilRep *domain.Representative
	do2 := toRepDO(nilRep)
	if do2.Name != "" {
		t.Errorf("expected empty for nil pointer")
	}
}

func TestToCorpDOAndBack(t *testing.T) {
	corp := &domain.Corporation{
		Name:               dp.CreateCorpName("Acme"),
		AllEmailDomains:    []string{"acme.com"},
		PrimaryEmailDomain: "acme.com",
	}
	do := toCorpDO(corp)
	if do.Name != "Acme" || do.Domain != "acme.com" {
		t.Errorf("toCorpDO mismatch")
	}
	result := do.toCorp()
	if result.Name.CorpName() != "Acme" {
		t.Errorf("toCorp mismatch")
	}
}

func TestToCorpDONil(t *testing.T) {
	do := toCorpDO(nil)
	if do.Name != "" || do.Domain != "" {
		t.Errorf("expected empty for nil")
	}
}

func TestIndividualSigningDO(t *testing.T) {
	is := &domain.IndividualSigning{
		Link: domain.LinkInfo{
			Id: "link-1",
			CLAInfo: domain.CLAInfo{
				CLAId:    "cla-1",
				Language: dp.CreateLanguage("en"),
			},
		},
		Rep: domain.Representative{
			Name:      dp.CreateName("Jane"),
			EmailAddr: dp.CreateEmailAddr("jane@test.com"),
		},
		Date:    "2024-01-01",
		AllInfo: domain.AllSingingInfo{"key": "val"},
		Logs: []domain.IndividualSigningLog{
			{Date: "2024-01-02", ClaId: "cla-2", Action: "update"},
		},
	}

	do := toIndividualSigningDO(is)
	if do.CLAId != "cla-1" || do.LinkId != "link-1" {
		t.Errorf("toIndividualSigningDO mismatch: %+v", do)
	}
	if do.Language != "en" {
		t.Errorf("expected 'en', got '%s'", do.Language)
	}

	result := do.toIndividualSigning()
	if result.Link.Id != "link-1" || len(result.Logs) != 1 {
		t.Errorf("toIndividualSigning mismatch: %+v", result)
	}
}

func TestIndividualSigningLogDO(t *testing.T) {
	log := domain.IndividualSigningLog{Date: "2024-01-01", ClaId: "cla-1", Action: "sign"}
	do := toIndividualSigningLogDO(log)
	if do.Date != "2024-01-01" {
		t.Errorf("expected '2024-01-01', got '%s'", do.Date)
	}
	result := do.toIndividualSigningLog()
	if result.ClaId != "cla-1" || result.Action != "sign" {
		t.Errorf("roundtrip failed: %+v", result)
	}
}

func TestToIndividualSigningLogsDO(t *testing.T) {
	logs := []domain.IndividualSigningLog{
		{Date: "d1", ClaId: "c1", Action: "a1"},
		{Date: "d2", ClaId: "c2", Action: "a2"},
	}
	do := toIndividualSigningLogsDO(logs)
	if len(do.Logs) != 2 {
		t.Errorf("expected 2, got %d", len(do.Logs))
	}
}

// --- Manager DO ---

func TestToManagerDOAndBack(t *testing.T) {
	m := &domain.Manager{
		Id: "manager-1",
		Representative: domain.Representative{
			Name:      dp.CreateName("Bob"),
			EmailAddr: dp.CreateEmailAddr("bob@test.com"),
		},
	}
	do := toManagerDO(m)
	if do.Id != "manager-1" || do.Name != "Bob" {
		t.Errorf("toManagerDO mismatch: %+v", do)
	}
	result := do.toManager()
	if result.Id != "manager-1" || result.Name.Name() != "Bob" {
		t.Errorf("toManager mismatch: %+v", result)
	}
}

func TestManagerDOIsEmpty(t *testing.T) {
	empty := managerDO{}
	if !empty.isEmpty() {
		t.Error("expected isEmpty=true for zero value")
	}
	full := managerDO{Id: "m1"}
	if full.isEmpty() {
		t.Error("expected isEmpty=false for non-empty")
	}
}

// --- Employee Signing DO ---

func TestToEmployeeSigningDOAndBack(t *testing.T) {
	es := &domain.EmployeeSigning{
		Id: "es-1",
		Rep: domain.Representative{
			Name:      dp.CreateName("Alice"),
			EmailAddr: dp.CreateEmailAddr("alice@test.com"),
		},
		CLA: domain.CLAInfo{
			CLAId:    "cla-1",
			Language: dp.CreateLanguage("en"),
		},
		Date:    "2024-01-01",
		Enabled: true,
		AllInfo: domain.AllSingingInfo{"k": "v"},
		Logs: []domain.EmployeeSigningLog{
			{Time: 100, Action: "sign"},
			{Time: 200, Action: "update"},
		},
	}
	do := toEmployeeSigningDO(es)
	if do.Id != "es-1" || do.CLAId != "cla-1" || !do.Enabled {
		t.Errorf("toEmployeeSigningDO mismatch: %+v", do)
	}
	result := do.toEmployeeSigning()
	if result.Id != "es-1" || !result.Enabled || len(result.Logs) != 2 {
		t.Errorf("toEmployeeSigning mismatch: %+v", result)
	}
}

func TestToEmployeeSigningLogDOs(t *testing.T) {
	logs := []domain.EmployeeSigningLog{
		{Time: 100, Action: "sign"},
		{Time: 200, Action: "update"},
	}
	dos := toEmployeeSigningLogDOs(logs)
	if len(dos) != 2 {
		t.Errorf("expected 2, got %d", len(dos))
	}
	if dos[0].Time != 100 || dos[0].Action != "sign" {
		t.Errorf("log[0] mismatch")
	}
	r := dos[0].toEmployeeSigningLog()
	if r.Time != 100 || r.Action != "sign" {
		t.Errorf("roundtrip mismatch")
	}
}

// --- Email Credential DO ---

func TestToEmailCredentialDOAndBack(t *testing.T) {
	e := &domain.EmailCredential{
		Addr:     dp.CreateEmailAddr("test@example.com"),
		Token:    []byte("secret"),
		Platform: "gmail",
	}
	do := toEmailCredentialDO(e)
	if do.Email != "test@example.com" || do.Platform != "gmail" {
		t.Errorf("toEmailCredentialDO mismatch: %+v", do)
	}
	result := do.toEmailCredential()
	if result.Addr.EmailAddr() != "test@example.com" || result.Platform != "gmail" {
		t.Errorf("toEmailCredential mismatch: %+v", result)
	}
}

// --- CorpSigning DO (more) ---

func TestToCorpSigningDO(t *testing.T) {
	cs := &domain.CorpSigning{
		Id:   "cs-1",
		Date: "2024-06-01",
		Link: domain.LinkInfo{
			Id: "link-1",
			CLAInfo: domain.CLAInfo{
				CLAId:    "cla-1",
				Language: dp.CreateLanguage("en"),
			},
		},
		Rep: domain.Representative{
			Name:      dp.CreateName("CEO"),
			EmailAddr: dp.CreateEmailAddr("ceo@corp.com"),
		},
		Corp: domain.Corporation{
			Name:               dp.CreateCorpName("BigCorp"),
			AllEmailDomains:    []string{"bigcorp.com"},
			PrimaryEmailDomain: "bigcorp.com",
		},
		AllInfo: domain.AllSingingInfo{"field": "value"},
		Admin: domain.Manager{
			Id: "admin-1",
			Representative: domain.Representative{
				Name:      dp.CreateName("Admin"),
				EmailAddr: dp.CreateEmailAddr("admin@bigcorp.com"),
			},
		},
		Managers: []domain.Manager{
			{
				Id: "mgr-1",
				Representative: domain.Representative{
					Name:      dp.CreateName("Mgr1"),
					EmailAddr: dp.CreateEmailAddr("mgr1@bigcorp.com"),
				},
			},
		},
	}

	do := toCorpSigningDO(cs)
	if do.CLAId != "cla-1" || do.LinkId != "link-1" || do.Language != "en" {
		t.Errorf("toCorpSigningDO mismatch: %+v", do)
	}
	if do.Rep.Name != "CEO" || do.Rep.Email != "ceo@corp.com" {
		t.Errorf("rep mismatch: %+v", do.Rep)
	}

	summary := do.toCorpSigningSummary()
	if summary.Id != do.index() {
		t.Errorf("expected '%s', got '%s'", do.index(), summary.Id)
	}
	if summary.Date != "2024-06-01" {
		t.Errorf("date mismatch")
	}

	result := do.toCorpSigning()
	if result.Id != do.index() || result.Link.Id != "link-1" {
		t.Errorf("toCorpSigning mismatch")
	}
}

func TestToCorpSigningDOForMigrate(t *testing.T) {
	cs := &domain.CorpSigning{
		Id:   "cs-mig",
		Date: "2024-01-01",
		Link: domain.LinkInfo{
			Id: "link-1",
			CLAInfo: domain.CLAInfo{
				CLAId:    "cla-1",
				Language: dp.CreateLanguage("en"),
			},
		},
		Rep: domain.Representative{
			Name:      dp.CreateName("CEO"),
			EmailAddr: dp.CreateEmailAddr("ceo@test.com"),
		},
		Corp: domain.Corporation{
			Name:               dp.CreateCorpName("Corp"),
			AllEmailDomains:    []string{"corp.com"},
			PrimaryEmailDomain: "corp.com",
		},
		Admin: domain.Manager{
			Id: "admin",
			Representative: domain.Representative{
				Name:      dp.CreateName("Admin"),
				EmailAddr: dp.CreateEmailAddr("admin@corp.com"),
			},
		},
	}

	do := toCorpSigningDOForMigrate(cs)
	if do.Admin.Name != "Admin" {
		t.Errorf("expected admin name 'Admin', got '%s'", do.Admin.Name)
	}
	if len(do.Managers) > 0 {
		t.Errorf("expected 0 managers, got %d", len(do.Managers))
	}
}

func TestAllManagers(t *testing.T) {
	do := &corpSigningDO{
		Admin: managerDO{
			Id: "admin",
			RepDO: RepDO{
				Name:  "Admin",
				Email: "admin@test.com",
			},
		},
		Managers: []managerDO{
			{Id: "mgr-1", RepDO: RepDO{Name: "Mgr1", Email: "mgr1@test.com"}},
		},
	}

	managers := do.allManagers()
	if len(managers) != 2 {
		t.Errorf("expected 2, got %d", len(managers))
	}
}

func TestToEmployeeSigningDOs(t *testing.T) {
	employees := []domain.EmployeeSigning{
		{
			Id: "es-1",
			Rep: domain.Representative{
				Name:      dp.CreateName("E1"),
				EmailAddr: dp.CreateEmailAddr("e1@test.com"),
			},
			CLA: domain.CLAInfo{CLAId: "cla-1", Language: dp.CreateLanguage("en")},
		},
	}
	dos := toEmployeeSigningDOs(employees)
	if len(dos) != 1 {
		t.Errorf("expected 1, got %d", len(dos))
	}
	if dos[0].Id != "es-1" {
		t.Errorf("expected 'es-1', got '%s'", dos[0].Id)
	}

	// Test nil
	nilDos := toEmployeeSigningDOs(nil)
	if nilDos != nil {
		t.Error("expected nil for nil input")
	}
}

func TestToManagerDOs(t *testing.T) {
	managers := []domain.Manager{
		{
			Id: "m1",
			Representative: domain.Representative{
				Name:      dp.CreateName("M1"),
				EmailAddr: dp.CreateEmailAddr("m1@test.com"),
			},
		},
	}
	dos := toManagerDOs(managers)
	if len(dos) != 1 || dos[0].Id != "m1" {
		t.Errorf("toManagerDOs mismatch")
	}

	// Test nil
	nilDos := toManagerDOs(nil)
	if nilDos != nil {
		t.Error("expected nil for nil input")
	}
}

func TestToManagers(t *testing.T) {
	do := &corpSigningDO{
		Managers: []managerDO{
			{Id: "m1", RepDO: RepDO{Name: "M1", Email: "m1@test.com"}},
		},
	}
	ms := do.toManagers()
	if len(ms) != 1 || ms[0].Id != "m1" {
		t.Errorf("toManagers mismatch")
	}
}

func TestToEmployeeSignings(t *testing.T) {
	do := &corpSigningDO{
		Employees: []employeeSigningDO{
			{Id: "es-1", CLAId: "cla-1"},
		},
	}
	ess := do.toEmployeeSignings()
	if len(ess) != 1 || ess[0].Id != "es-1" {
		t.Errorf("toEmployeeSignings mismatch")
	}
}

func TestConstValues(t *testing.T) {
	// Ensure important constants are defined
	if fieldPDF == "" || fieldRep == "" || fieldId == "" {
		t.Error("expected constant definitions")
	}
}

func TestToUserDO(t *testing.T) {
	u := &domain.User{
		LinkId:        "link-1",
		CorpSigningId: "cs-1",
		UserBasicInfo: domain.UserBasicInfo{
			Id:              "user-1",
			Account:         dp.CreateAccount("user001"),
			Password:        []byte("enc"),
			EmailAddr:       dp.CreateEmailAddr("user@test.com"),
			PasswordChanged: true,
			PrivacyConsent:  domain.PrivacyConsent{Time: "2024-01-01", Version: "v1"},
			Version:         1,
		},
	}
	do := toUserDO(u)
	if do.Email != "user@test.com" || do.Account != "user001" || do.LinkId != "link-1" {
		t.Errorf("toUserDO mismatch: %+v", do)
	}

	result := do.toUser()
	if result.LinkId != "link-1" || result.UserBasicInfo.Id == "" || result.Account.Account() != "user001" {
		t.Errorf("toUser mismatch: %+v", result)
	}
}

func TestToUserDOForMigrate(t *testing.T) {
	u := &domain.User{
		LinkId:        "link-1",
		CorpSigningId: "cs-1",
		UserBasicInfo: domain.UserBasicInfo{
			EmailAddr:      dp.CreateEmailAddr("user@test.com"),
			Account:        dp.CreateAccount("user001"),
			PrivacyConsent: domain.PrivacyConsent{Time: "2024-01-01", Version: "v2"},
		},
	}
	do := toUserDOForMigrate(u)
	if do.Email != "user@test.com" || do.PrivacyConsent.Version != "v2" {
		t.Errorf("toUserDOForMigrate mismatch: %+v", do)
	}

	// nil test
	nilDo := toUserDOForMigrate(nil)
	if nilDo.Email != "" {
		t.Errorf("expected empty for nil")
	}
}

func TestToPrivacyConsentDO(t *testing.T) {
	p := domain.PrivacyConsent{Time: "2024-01-01", Version: "v1"}
	do := toPrivacyConsentDO(p)
	if do.Time != "2024-01-01" || do.Version != "v1" {
		t.Errorf("toPrivacyConsentDO mismatch: %+v", do)
	}
}

func TestToVerificationCodeDOAndBack(t *testing.T) {
	purpose, _ := dp.NewPurpose("test-purpose")
	vc := &domain.VerificationCode{
		Expiry: 100,
		VerificationCodeKey: domain.VerificationCodeKey{
			Code:    "123456",
			Purpose: purpose,
		},
	}
	do := toVerificationCodeDO(vc)
	if do.Code != "123456" || do.Expiry != 100 || do.Purpose != "test-purpose" {
		t.Errorf("toVerificationCodeDO mismatch: %+v", do)
	}
	result := do.toVerificationCode()
	if result.Code != "123456" || result.Expiry != 100 {
		t.Errorf("toVerificationCode mismatch: %+v", result)
	}
}

func TestToVerificationCodeFilter(t *testing.T) {
	purpose, _ := dp.NewPurpose("signing-purpose")
	key := &domain.VerificationCodeKey{
		Code:    "654321",
		Purpose: purpose,
	}
	filter := toVerificationCodeFilter(key)
	if filter[fieldCode] != "654321" || filter[fieldPurpose] != "signing-purpose" {
		t.Errorf("filter mismatch: %+v", filter)
	}
}

func TestPrivacyConsentDO(t *testing.T) {
	do := privacyConsentDO{Time: "2024-01-01", Version: "v1"}
	if do.Time != "2024-01-01" {
		t.Errorf("privacyConsentDO mismatch")
	}
}
