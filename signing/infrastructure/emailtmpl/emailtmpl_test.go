package emailtmpl

import (
	"text/template"
	"testing"
)

func TestFindTmpl(t *testing.T) {
	// Save original and restore
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplCorporationSigning: template.Must(template.New("test").Parse("Hello {{.Org}}")),
	}
	defer func() { msgTmpl = orig }()

	tmpl := findTmpl(TmplCorporationSigning)
	if tmpl == nil {
		t.Error("expected to find template")
	}

	tmpl = findTmpl("nonexistent")
	if tmpl != nil {
		t.Error("expected nil for nonexistent template")
	}
}

func TestGenEmailMsg(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplCorporationSigning: template.Must(template.New("test").Parse("Org: {{.Org}}")),
	}
	defer func() { msgTmpl = orig }()

	msg, err := genEmailMsg(TmplCorporationSigning, struct{ Org string }{Org: "TestOrg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content.String() != "Org: TestOrg" {
		t.Errorf("expected 'Org: TestOrg', got '%s'", msg.Content.String())
	}
}

func TestGenEmailMsgMissingTemplate(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{}
	defer func() { msgTmpl = orig }()

	_, err := genEmailMsg("nonexistent", nil)
	if err == nil {
		t.Error("expected error for missing template")
	}
}

func TestCorporationSigningGenEmailMsg(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplCorporationSigning: template.Must(template.New("test").Parse("{{.Org}} - {{.Date}}")),
	}
	defer func() { msgTmpl = orig }()

	data := &CorporationSigning{Org: "Org1", Date: "2024-01-01"}
	msg, err := data.GenEmailMsg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content.String() != "Org1 - 2024-01-01" {
		t.Errorf("unexpected content: %s", msg.Content.String())
	}
}

func TestIndividualSigningGenEmailMsg(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplIndividualSigning: template.Must(template.New("test").Parse("Name: {{.Name}}")),
	}
	defer func() { msgTmpl = orig }()

	data := &IndividualSigning{Name: "John"}
	msg, err := data.GenEmailMsg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content.String() != "Name: John" {
		t.Errorf("unexpected content: %s", msg.Content.String())
	}
}

func TestVerificationCodeGenEmailMsg(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplVerificationCode: template.Must(template.New("test").Parse("{{.Code}}")),
	}
	defer func() { msgTmpl = orig }()

	data := &VerificationCode{Code: "123456"}
	msg, err := data.GenEmailMsg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content.String() != "123456" {
		t.Errorf("unexpected content: %s", msg.Content.String())
	}
}

func TestAddingCorpManagerAdmin(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplAddingCorpAdmin: template.Must(template.New("test").Parse("Admin: {{.User}}")),
	}
	defer func() { msgTmpl = orig }()

	data := &AddingCorpManager{Admin: true, User: "AdminUser"}
	msg, err := data.GenEmailMsg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !msg.HasSecret {
		t.Error("expected HasSecret=true for admin")
	}
}

func TestAddingCorpManagerNonAdmin(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplAddingCorpManager: template.Must(template.New("test").Parse("Manager: {{.User}}")),
	}
	defer func() { msgTmpl = orig }()

	data := &AddingCorpManager{Admin: false, User: "MgrUser"}
	msg, err := data.GenEmailMsg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !msg.HasSecret {
		t.Error("expected HasSecret=true for manager")
	}
}

func TestEmployeeNotificationActive(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplActivatingEmployee: template.Must(template.New("test").Parse("Active: {{.Name}}")),
	}
	defer func() { msgTmpl = orig }()

	data := &EmployeeNotification{Active: true, Name: "Emp1"}
	msg, err := data.GenEmailMsg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content.String() != "Active: Emp1" {
		t.Errorf("unexpected content: %s", msg.Content.String())
	}
}

func TestEmployeeNotificationInactive(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplInactivaingEmployee: template.Must(template.New("test").Parse("Inactive: {{.Name}}")),
	}
	defer func() { msgTmpl = orig }()

	data := &EmployeeNotification{Inactive: true, Name: "Emp2"}
	msg, err := data.GenEmailMsg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content.String() != "Inactive: Emp2" {
		t.Errorf("unexpected content: %s", msg.Content.String())
	}
}

func TestEmployeeNotificationRemoving(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplRemovingingEmployee: template.Must(template.New("test").Parse("Removed: {{.Name}}")),
	}
	defer func() { msgTmpl = orig }()

	data := &EmployeeNotification{Removing: true, Name: "Emp3"}
	msg, err := data.GenEmailMsg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content.String() != "Removed: Emp3" {
		t.Errorf("unexpected content: %s", msg.Content.String())
	}
}

func TestEmployeeNotificationNoAction(t *testing.T) {
	data := &EmployeeNotification{}
	_, err := data.GenEmailMsg()
	if err == nil {
		t.Error("expected error for no action")
	}
}

func TestPasswordRetrievalGenEmailMsg(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplPasswordRetrieval: template.Must(template.New("test").Parse("Reset: {{.ResetURL}}")),
	}
	defer func() { msgTmpl = orig }()

	data := PasswordRetrieval{ResetURL: "https://reset.example.com"}
	msg, err := data.GenEmailMsg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.MIME == "" {
		t.Error("expected MIME to be set for password retrieval")
	}
	if msg.Content.String() != "Reset: https://reset.example.com" {
		t.Errorf("unexpected content: %s", msg.Content.String())
	}
}

func TestCLAUpdatedGenEmailMsg(t *testing.T) {
	orig := msgTmpl
	msgTmpl = map[string]*template.Template{
		TmplCLAUpdated: template.Must(template.New("test").Parse("CLA: {{.Org}}")),
	}
	defer func() { msgTmpl = orig }()

	data := &CLAUpdated{Org: "TestOrg"}
	msg, err := data.GenEmailMsg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content.String() != "CLA: TestOrg" {
		t.Errorf("unexpected content: %s", msg.Content.String())
	}
}

func TestAllTemplateConstants(t *testing.T) {
	names := []string{
		TmplCorporationSigning, TmplIndividualSigning, TmplEmployeeSigning,
		TmplNotifyingManager, TmplVerificationCode, TmplAddingCorpEmailDomain,
		TmplAddingCorpAdmin, TmplAddingCorpManager, TmplRemovingCorpManager,
		TmplActivatingEmployee, TmplInactivaingEmployee, TmplRemovingingEmployee,
		TmplPasswordRetrieval, TmplEmailVerification, TmplCLAUpdated,
	}
	for _, name := range names {
		if name == "" {
			t.Error("empty template name constant")
		}
	}
}
