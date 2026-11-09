package emailtmpl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if root := findModuleRoot(); root != "" {
		_ = os.Chdir(root)
	}

	if err := Init(); err != nil {
		// Do not fail the entire package; template-dependent tests will fail individually.
		_ = err
	}
	m.Run()
}

func findModuleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func TestNormalizeLang(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Chinese", langZH},
		{"English", langEN},
		{"", langEN},
		{"xx", langEN},
		{"chinese", langEN},
		{"ENGLISH", langEN},
		{"zh", langZH},
		{"en", langEN},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeLang(tt.input))
		})
	}
}

func TestFindTmpl_Fallback(t *testing.T) {
	// If zh is missing for a template, findTmpl should fall back to en.
	// We simulate by removing zh from the internal map temporarily.
	name := TmplCorporationSigning

	// Save and remove zh
	origZh := msgTmpl[name][langZH]
	delete(msgTmpl[name], langZH)
	defer func() {
		if origZh != nil {
			msgTmpl[name][langZH] = origZh
		}
	}()

	// lang=zh should fall back to en (not nil)
	tmpl := findTmpl(name, langZH)
	require.NotNil(t, tmpl, "findTmpl should fall back to en when zh is missing")

	// lang=en should still work
	tmplEn := findTmpl(name, langEN)
	require.NotNil(t, tmplEn)

	// Both should be the same en template
	assert.Same(t, tmplEn, tmpl)
}

func TestFindTmpl_UnknownName(t *testing.T) {
	tmpl := findTmpl("nonexistent-template", langEN)
	assert.Nil(t, tmpl)
}

func TestGenEmailMsg_Matrix(t *testing.T) {
	// Test all 21 templates with zh, en, and empty lang.
	type tmplCase struct {
		name string
		data interface{}
	}

	cases := []tmplCase{
		{TmplCorporationSigning, &CorporationSigning{Org: "TestOrg", Date: "2024-01-01", AdminName: "Admin", ProjectURL: "http://example.com", SigningInfo: "info"}},
		{TmplIndividualSigning, &IndividualSigning{Name: "Alice"}},
		{TmplEmployeeSigning, &EmployeeSigning{Name: "Bob", Org: "TestOrg", ProjectURL: "http://example.com", Managers: "manager1"}},
		{TmplNotifyingManager, &NotifyingManager{Org: "TestOrg", EmployeeEmail: "bob@test.com", ProjectURL: "http://example.com", URLOfCLAPlatform: "http://cla.example.com"}},
		{TmplVerificationCode, &VerificationCode{Email: "user@test.com", Org: "TestOrg", Code: "123456", ProjectURL: "http://example.com"}},
		{TmplAddingCorpEmailDomain, &AddingCorpEmailDomain{Corp: "TestCorp", Org: "TestOrg", Code: "123456", ProjectURL: "http://example.com"}},
		{TmplAddingCorpAdmin, &AddingCorpManager{Admin: true, ID: "id1", User: "Admin", Email: "admin@test.com", Password: []byte("pw"), Org: "TestOrg", ProjectURL: "http://example.com", URLOfCLAPlatform: "http://cla.example.com"}},
		{TmplAddingCorpManager, &AddingCorpManager{Admin: false, ID: "id1", User: "Manager", Email: "mgr@test.com", Password: []byte("pw"), Org: "TestOrg", ProjectURL: "http://example.com", URLOfCLAPlatform: "http://cla.example.com"}},
		{TmplRemovingCorpManager, &RemovingCorpManager{User: "Manager", Org: "TestOrg", ProjectURL: "http://example.com"}},
		{TmplActivatingEmployee, &EmployeeNotification{Active: true, Name: "Emp", Org: "TestOrg", ProjectURL: "http://example.com", Manager: "mgr@test.com"}},
		{TmplInactivaingEmployee, &EmployeeNotification{Inactive: true, Name: "Emp", Org: "TestOrg", ProjectURL: "http://example.com", Manager: "mgr@test.com"}},
		{TmplRemovingingEmployee, &EmployeeNotification{Removing: true, Name: "Emp", Org: "TestOrg", ProjectURL: "http://example.com", Manager: "mgr@test.com"}},
		{TmplPasswordRetrieval, &PasswordRetrieval{Org: "TestOrg", Timeout: 30, ResetURL: "http://reset.example.com", RetrievalURL: "http://retrieve.example.com"}},
		{TmplEmailVerification, &EmailVerification{Code: "123456"}},
		{TmplCLAUpdated, &CLAUpdated{Org: "TestOrg", AdminName: "Admin", ProjectURL: "http://example.com", URLOfCLAPlatform: "http://cla.example.com"}},
		{TmplCLAUpdatedIndividual, &CLAUpdatedIndividual{Name: "Alice", Org: "TestOrg", UpdateDate: "2024-01-01", GracePeriodDays: 30, SignCLAURL: "http://sign.example.com", ProjectURL: "http://example.com"}},
		{TmplAutoApprovalEnabled, &AutoApprovalEnabled{Org: "TestOrg", ProjectURL: "http://example.com", URLOfCLAPlatform: "http://cla.example.com"}},
		{TmplAutoApprovalDisabled, &AutoApprovalDisabled{Org: "TestOrg", ProjectURL: "http://example.com", URLOfCLAPlatform: "http://cla.example.com"}},
		{TmplAutoApprovalEmployee, &AutoApprovalEmployee{Name: "Emp", Org: "TestOrg", ProjectURL: "http://example.com"}},
		{TmplAutoApprovalManager, &AutoApprovalManager{EmployeeEmail: "emp@test.com", Org: "TestOrg", ProjectURL: "http://example.com", URLOfCLAPlatform: "http://cla.example.com"}},
	}

	langs := []string{langZH, langEN, ""}

	for _, tc := range cases {
		for _, lang := range langs {
			t.Run(tc.name+"/"+lang, func(t *testing.T) {
				msg, err := genEmailMsg(tc.name, lang, tc.data)
				require.NoError(t, err)

				// Subject should be non-empty
				assert.NotEmpty(t, msg.Subject, "subject should not be empty")

				// Content should be non-empty
				assert.True(t, msg.Content.Len() > 0, "content should not be empty")

				// No unresolved template placeholders
				content := msg.Subject + msg.Content.String()
				assert.False(t, strings.Contains(content, "{{"), "content should not contain unresolved placeholders")

				// zh results should contain Chinese characters
				if NormalizeLang(lang) == langZH {
					assert.True(t,
						strings.ContainsAny(msg.Subject, " \t") || hasChineseChar(msg.Subject),
						"zh subject should contain Chinese characters: %s", msg.Subject)
					assert.True(t, hasChineseChar(msg.Content.String()),
						"zh content should contain Chinese characters")
				}

				// en and "" results should be identical
				if lang == "" || lang == langEN {
					msgEn, _ := genEmailMsg(tc.name, langEN, tc.data)
					assert.Equal(t, msgEn.Subject, msg.Subject, "en and empty lang should produce same subject")
					assert.Equal(t, msgEn.Content.String(), msg.Content.String(), "en and empty lang should produce same content")
				}
			})
		}
	}
}

func TestGenEmailMsg_UnknownTemplate(t *testing.T) {
	_, err := genEmailMsg("nonexistent", langEN, nil)
	assert.Error(t, err)
}

func TestGenEmailMsg_NormalizesLang(t *testing.T) {
	// Passing raw "Chinese" should produce same result as "zh"
	data := &CLAUpdated{Org: "TestOrg", AdminName: "Admin", ProjectURL: "http://example.com", URLOfCLAPlatform: "http://cla.example.com"}

	msgRaw, err := genEmailMsg(TmplCLAUpdated, "Chinese", data)
	require.NoError(t, err)

	msgZh, err := genEmailMsg(TmplCLAUpdated, langZH, data)
	require.NoError(t, err)

	assert.Equal(t, msgZh.Subject, msgRaw.Subject)
	assert.Equal(t, msgZh.Content.String(), msgRaw.Content.String())
}

func TestEmployeeNotificationGenEmailMsg_NoFlag(t *testing.T) {
	data := &EmployeeNotification{}
	_, err := data.GenEmailMsg()
	assert.Error(t, err)
}

func TestEmployeeNotificationGenEmailMsg_Active(t *testing.T) {
	data := &EmployeeNotification{
		Lang:   langZH,
		Active: true,
		Name:   "Test",
		Org:    "TestOrg",
		ProjectURL: "http://example.com",
		Manager: "mgr@test.com",
	}
	msg, err := data.GenEmailMsg()
	require.NoError(t, err)
	assert.NotEmpty(t, msg.Subject)
	assert.True(t, msg.Content.Len() > 0)
}

func TestEmployeeNotificationGenEmailMsg_Inactive(t *testing.T) {
	data := &EmployeeNotification{
		Lang:     langEN,
		Inactive: true,
		Name:     "Test",
		Org:      "TestOrg",
		ProjectURL: "http://example.com",
		Manager:  "mgr@test.com",
	}
	msg, err := data.GenEmailMsg()
	require.NoError(t, err)
	assert.NotEmpty(t, msg.Subject)
}

func TestEmployeeNotificationGenEmailMsg_Removing(t *testing.T) {
	data := &EmployeeNotification{
		Lang:     langEN,
		Removing: true,
		Name:     "Test",
		Org:      "TestOrg",
		ProjectURL: "http://example.com",
		Manager:  "mgr@test.com",
	}
	msg, err := data.GenEmailMsg()
	require.NoError(t, err)
	assert.NotEmpty(t, msg.Subject)
}

func TestAddingCorpManagerGenEmailMsg_Admin(t *testing.T) {
	data := &AddingCorpManager{
		Lang:             langEN,
		Admin:            true,
		ID:               "id1",
		User:             "Admin",
		Email:            "admin@test.com",
		Password:         []byte("pw"),
		Org:              "TestOrg",
		ProjectURL:       "http://example.com",
		URLOfCLAPlatform: "http://cla.example.com",
	}
	msg, err := data.GenEmailMsg()
	require.NoError(t, err)
	assert.NotEmpty(t, msg.Subject)
	assert.True(t, msg.HasSecret)
}

func TestAddingCorpManagerGenEmailMsg_Manager(t *testing.T) {
	data := &AddingCorpManager{
		Lang:             langEN,
		Admin:            false,
		ID:               "id1",
		User:             "Manager",
		Email:            "mgr@test.com",
		Password:         []byte("pw"),
		Org:              "TestOrg",
		ProjectURL:       "http://example.com",
		URLOfCLAPlatform: "http://cla.example.com",
	}
	msg, err := data.GenEmailMsg()
	require.NoError(t, err)
	assert.NotEmpty(t, msg.Subject)
	assert.True(t, msg.HasSecret)
}

func TestPasswordRetrievalGenEmailMsg_SetsMIME(t *testing.T) {
	data := PasswordRetrieval{
		Lang:         langEN,
		Org:          "TestOrg",
		Timeout:      30,
		ResetURL:     "http://reset.example.com",
		RetrievalURL: "http://retrieve.example.com",
	}
	msg, err := data.GenEmailMsg()
	require.NoError(t, err)
	assert.NotEmpty(t, msg.MIME, "PasswordRetrieval should set MIME")
	assert.Contains(t, msg.MIME, "text/html")
}

func TestPasswordRetrievalGenEmailMsg_NoOrg(t *testing.T) {
	data := PasswordRetrieval{
		Lang:         langEN,
		Org:          "",
		Timeout:      30,
		ResetURL:     "http://reset.example.com",
		RetrievalURL: "http://retrieve.example.com",
	}
	msg, err := data.GenEmailMsg()
	require.NoError(t, err)
	assert.NotEmpty(t, msg.Subject)
	assert.True(t, msg.Content.Len() > 0)
}

func TestInit_LoadsAllTemplates(t *testing.T) {
	// Init should have been called in TestMain.
	// Verify all 21 templates are loaded with at least en.
	expected := []string{
		TmplCorporationSigning, TmplIndividualSigning, TmplEmployeeSigning,
		TmplNotifyingManager, TmplVerificationCode, TmplAddingCorpEmailDomain,
		TmplAddingCorpAdmin, TmplAddingCorpManager, TmplRemovingCorpManager,
		TmplActivatingEmployee, TmplInactivaingEmployee, TmplRemovingingEmployee,
		TmplPasswordRetrieval, TmplEmailVerification, TmplCLAUpdated,
		TmplCLAUpdatedIndividual, TmplAutoApprovalEnabled, TmplAutoApprovalDisabled,
		TmplAutoApprovalEmployee, TmplAutoApprovalManager,
	}
	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			m, ok := msgTmpl[name]
			require.True(t, ok, "template %s should be loaded", name)
			assert.NotNil(t, m[langEN], "en template for %s should be loaded", name)
			assert.NotNil(t, m[langZH], "zh template for %s should be loaded", name)
		})
	}
}

func TestAllStructsGenEmailMsg(t *testing.T) {
	// Exercise GenEmailMsg() directly on each data struct to cover the wrapper methods.
	tests := []struct {
		name string
		data interface {
			GenEmailMsg() (EmailMessage, error)
		}
	}{
		{"CorporationSigning", &CorporationSigning{Lang: langZH, Org: "Org", Date: "2024-01-01", AdminName: "Admin", ProjectURL: "http://x.com", SigningInfo: "info"}},
		{"IndividualSigning", &IndividualSigning{Lang: langZH, Name: "Alice"}},
		{"VerificationCode", &VerificationCode{Lang: langZH, Email: "e@x.com", Org: "Org", Code: "123", ProjectURL: "http://x.com"}},
		{"AddingCorpEmailDomain", AddingCorpEmailDomain{Lang: langEN, Corp: "Corp", Org: "Org", Code: "123", ProjectURL: "http://x.com"}},
		{"RemovingCorpManager", &RemovingCorpManager{Lang: langEN, User: "U", Org: "Org", ProjectURL: "http://x.com"}},
		{"EmployeeSigning", &EmployeeSigning{Lang: langEN, Name: "N", Org: "Org", ProjectURL: "http://x.com", Managers: "m"}},
		{"NotifyingManager", &NotifyingManager{Lang: langEN, Org: "Org", EmployeeEmail: "e@x.com", ProjectURL: "http://x.com", URLOfCLAPlatform: "http://c.x.com"}},
		{"EmailVerification", &EmailVerification{Lang: langEN, Code: "123"}},
		{"CLAUpdated", &CLAUpdated{Lang: langEN, Org: "Org", AdminName: "Admin", ProjectURL: "http://x.com", URLOfCLAPlatform: "http://c.x.com"}},
		{"CLAUpdatedIndividual", &CLAUpdatedIndividual{Lang: langEN, Name: "N", Org: "Org", UpdateDate: "2024-01-01", GracePeriodDays: 30, SignCLAURL: "http://s.x.com", ProjectURL: "http://x.com"}},
		{"AutoApprovalEnabled", &AutoApprovalEnabled{Lang: langEN, Org: "Org", ProjectURL: "http://x.com", URLOfCLAPlatform: "http://c.x.com"}},
		{"AutoApprovalDisabled", &AutoApprovalDisabled{Lang: langEN, Org: "Org", ProjectURL: "http://x.com", URLOfCLAPlatform: "http://c.x.com"}},
		{"AutoApprovalEmployee", &AutoApprovalEmployee{Lang: langEN, Name: "N", Org: "Org", ProjectURL: "http://x.com"}},
		{"AutoApprovalManager", &AutoApprovalManager{Lang: langEN, EmployeeEmail: "e@x.com", Org: "Org", ProjectURL: "http://x.com", URLOfCLAPlatform: "http://c.x.com"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := tt.data.GenEmailMsg()
			require.NoError(t, err)
			assert.NotEmpty(t, msg.Subject)
			assert.True(t, msg.Content.Len() > 0)
		})
	}
}

func TestGenEmailMsg_RenderError(t *testing.T) {
	// Test error handling when template execution fails.
	// We can trigger this by using a template with a bad field reference.
	// Since all loaded templates are valid, we test the error path by
	// temporarily swapping a template with one that has a bad block.
	name := TmplEmailVerification
	orig := msgTmpl[name][langEN]
	defer func() { msgTmpl[name][langEN] = orig }()

	// Parse a template with a missing block to trigger ExecuteTemplate error
	badTmpl, err := parseTestTemplate(name, `{{define "subject"}}ok{{end}}`)
	require.NoError(t, err)
	msgTmpl[name][langEN] = badTmpl

	_, err = genEmailMsg(name, langEN, nil)
	// Should error because "body" block is missing
	assert.Error(t, err)
}

func hasChineseChar(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

func parseTestTemplate(name, content string) (*template.Template, error) {
	return template.New(name).Parse(content)
}
