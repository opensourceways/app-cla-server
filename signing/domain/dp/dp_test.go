package dp

import (
	"testing"
)

func init() {
	config = Config{
		MaxLengthOfName:     100,
		MaxLengthOfTitle:    200,
		MaxLengthOfEmail:    100,
		MaxLengthOfAccount:  50,
		MaxLengthOfCorpName: 100,
		CLA: []claConfig{
			{
				Type: "corporation",
				Fileds: []claFields{
					{
						Language: "en",
						Fileds:   []CLAField{{Type: "text", Desc: "Name", Title: "Name", MaxLength: 50}},
					},
				},
			},
			{
				Type: "individual",
				Fileds: []claFields{
					{
						Language: "en",
						Fileds:   []CLAField{{Type: "text", Desc: "Email", Title: "Email", MaxLength: 50}},
					},
				},
			},
		},
	}
	config.Validate()
}

// --- Purpose ---

func TestNewPurpose(t *testing.T) {
	p, err := NewPurpose("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Purpose() != "test" {
		t.Errorf("expected 'test', got '%s'", p.Purpose())
	}
}

func TestNewPurposeEmpty(t *testing.T) {
	_, err := NewPurpose("")
	if err == nil {
		t.Error("expected error for empty purpose")
	}
}

func TestCreatePurpose(t *testing.T) {
	p := CreatePurpose("any")
	if p.Purpose() != "any" {
		t.Errorf("expected 'any', got '%s'", p.Purpose())
	}
}

// --- URL ---

func TestNewURL(t *testing.T) {
	u, err := NewURL("https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.URL() != "https://example.com" {
		t.Errorf("expected 'https://example.com', got '%s'", u.URL())
	}
}

func TestNewURLEmpty(t *testing.T) {
	_, err := NewURL("")
	if err == nil {
		t.Error("expected error for empty url")
	}
}

func TestNewURLInvalid(t *testing.T) {
	_, err := NewURL("not a url")
	if err == nil {
		t.Error("expected error for invalid url")
	}
}

func TestCreateURL(t *testing.T) {
	u := CreateURL("any")
	if u.URL() != "any" {
		t.Errorf("expected 'any', got '%s'", u.URL())
	}
}

// --- CLAType ---

func TestNewCLATypeCorp(t *testing.T) {
	ct, err := NewCLAType("corporation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct.CLAType() != "corporation" {
		t.Errorf("expected 'corporation', got '%s'", ct.CLAType())
	}
}

func TestNewCLATypeIndividual(t *testing.T) {
	ct, err := NewCLAType("individual")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct.CLAType() != "individual" {
		t.Errorf("expected 'individual', got '%s'", ct.CLAType())
	}
}

func TestNewCLATypeInvalid(t *testing.T) {
	_, err := NewCLAType("unknown")
	if err == nil {
		t.Error("expected error for invalid cla type")
	}
}

func TestCreateCLAType(t *testing.T) {
	ct := CreateCLAType("any")
	if ct.CLAType() != "any" {
		t.Errorf("expected 'any', got '%s'", ct.CLAType())
	}
}

func TestIsCLATypeIndividual(t *testing.T) {
	if !IsCLATypeIndividual(CLATypeIndividual) {
		t.Error("expected true for individual type")
	}
	if IsCLATypeIndividual(CLATypeCorp) {
		t.Error("expected false for corp type")
	}
	if IsCLATypeIndividual(nil) {
		t.Error("expected false for nil")
	}
}

// --- Name ---

func TestNewName(t *testing.T) {
	n, err := NewName("John")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Name() != "John" {
		t.Errorf("expected 'John', got '%s'", n.Name())
	}
}

func TestNewNameEmpty(t *testing.T) {
	_, err := NewName("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestNewNameXSS(t *testing.T) {
	_, err := NewName("<script>")
	if err == nil {
		t.Error("expected error for XSS content")
	}
}

func TestNewNameTooLong(t *testing.T) {
	s := make([]byte, config.MaxLengthOfName+1)
	for i := range s {
		s[i] = 'a'
	}
	_, err := NewName(string(s))
	if err == nil {
		t.Error("expected error for long name")
	}
}

func TestCreateName(t *testing.T) {
	n := CreateName("any")
	if n.Name() != "any" {
		t.Errorf("expected 'any', got '%s'", n.Name())
	}
}

// --- Title ---

func TestNewTitle(t *testing.T) {
	title, err := NewTitle("Engineer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title.Title() != "Engineer" {
		t.Errorf("expected 'Engineer', got '%s'", title.Title())
	}
}

func TestNewTitleEmpty(t *testing.T) {
	_, err := NewTitle("")
	if err == nil {
		t.Error("expected error for empty title")
	}
}

func TestNewTitleTooLong(t *testing.T) {
	s := make([]byte, config.MaxLengthOfTitle+1)
	for i := range s {
		s[i] = 'a'
	}
	_, err := NewTitle(string(s))
	if err == nil {
		t.Error("expected error for long title")
	}
}

// --- EmailAddr ---

func TestNewEmailAddr(t *testing.T) {
	e, err := NewEmailAddr("test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.EmailAddr() != "test@example.com" {
		t.Errorf("expected 'test@example.com', got '%s'", e.EmailAddr())
	}
	if e.Domain() != "example.com" {
		t.Errorf("expected 'example.com', got '%s'", e.Domain())
	}
}

func TestNewEmailAddrInvalid(t *testing.T) {
	_, err := NewEmailAddr("not-an-email")
	if err == nil {
		t.Error("expected error for invalid email")
	}
}

func TestNewEmailAddrEmpty(t *testing.T) {
	_, err := NewEmailAddr("")
	if err == nil {
		t.Error("expected error for empty email")
	}
}

func TestCreateEmailAddr(t *testing.T) {
	e := CreateEmailAddr("any@test.com")
	if e.EmailAddr() != "any@test.com" {
		t.Errorf("expected 'any@test.com', got '%s'", e.EmailAddr())
	}
}

// --- Account ---

func TestNewAccountCommunity(t *testing.T) {
	a, err := NewAccount("valid-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Account() != "valid-user" {
		t.Errorf("expected 'valid-user', got '%s'", a.Account())
	}
}

func TestNewAccountCorpManager(t *testing.T) {
	a, err := NewAccount("valid_user-org.enterprise.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Account() != "valid_user-org.enterprise.com" {
		t.Errorf("unexpected account value")
	}
}

func TestNewAccountInvalid(t *testing.T) {
	_, err := NewAccount("invalid account with spaces")
	if err == nil {
		t.Error("expected error for invalid account")
	}
}

func TestNewAccountEmpty(t *testing.T) {
	_, err := NewAccount("")
	if err == nil {
		t.Error("expected error for empty account")
	}
}

func TestNewAccountTooLong(t *testing.T) {
	s := make([]byte, config.MaxLengthOfAccount+1)
	for i := range s {
		s[i] = 'a'
	}
	_, err := NewAccount(string(s))
	if err == nil {
		t.Error("expected error for long account")
	}
}

func TestCreateAccount(t *testing.T) {
	a := CreateAccount("any")
	if a.Account() != "any" {
		t.Errorf("expected 'any', got '%s'", a.Account())
	}
}

// --- CorpName ---

func TestNewCorpName(t *testing.T) {
	n, err := NewCorpName("Acme Corp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.CorpName() != "Acme Corp" {
		t.Errorf("expected 'Acme Corp', got '%s'", n.CorpName())
	}
}

func TestNewCorpNameEmpty(t *testing.T) {
	_, err := NewCorpName("")
	if err == nil {
		t.Error("expected error for empty corp name")
	}
}

func TestNewCorpNameXSS(t *testing.T) {
	_, err := NewCorpName("<script>alert(1)</script>")
	if err == nil {
		t.Error("expected error for XSS in corp name")
	}
}

func TestNewCorpNameTooLong(t *testing.T) {
	s := make([]byte, config.MaxLengthOfCorpName+1)
	for i := range s {
		s[i] = 'a'
	}
	_, err := NewCorpName(string(s))
	if err == nil {
		t.Error("expected error for long corp name")
	}
}

func TestCreateCorpName(t *testing.T) {
	n := CreateCorpName("any")
	if n.CorpName() != "any" {
		t.Errorf("expected 'any', got '%s'", n.CorpName())
	}
}

// --- Password ---

func TestNewPassword(t *testing.T) {
	p, err := NewPassword([]byte("secret"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(p.Password()) != "secret" {
		t.Errorf("expected 'secret', got '%s'", string(p.Password()))
	}
}

func TestNewPasswordEmpty(t *testing.T) {
	_, err := NewPassword([]byte{})
	if err == nil {
		t.Error("expected error for empty password")
	}
}

func TestPasswordClear(t *testing.T) {
	p, _ := NewPassword([]byte("secret"))
	p.Clear()
	b := p.Password()
	for _, v := range b {
		if v != 0 {
			t.Error("expected all bytes to be zero after clear")
		}
	}
}

func TestIsSamePassword(t *testing.T) {
	p1, _ := NewPassword([]byte("secret"))
	p2, _ := NewPassword([]byte("secret"))
	p3, _ := NewPassword([]byte("other"))
	p4, _ := NewPassword([]byte("secret!"))

	if !IsSamePassword(p1, p2) {
		t.Error("expected same password")
	}
	if IsSamePassword(p1, p3) {
		t.Error("expected different password")
	}
	if IsSamePassword(p1, p4) {
		t.Error("expected different password (different length)")
	}
}

// --- Language ---

func TestNewLanguage(t *testing.T) {
	l, err := NewLanguage("en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Language() != "en" {
		t.Errorf("expected 'en', got '%s'", l.Language())
	}
}

func TestNewLanguageInvalid(t *testing.T) {
	_, err := NewLanguage("fr")
	if err == nil {
		t.Error("expected error for unsupported language")
	}
}

func TestCreateLanguage(t *testing.T) {
	l := CreateLanguage("any")
	if l.Language() != "any" {
		t.Errorf("expected 'any', got '%s'", l.Language())
	}
}

// --- Config ---

func TestConfigValidate(t *testing.T) {
	cfg := &Config{
		MaxLengthOfName:  50,
		MaxLengthOfEmail: 50,
		CLA: []claConfig{
			{Type: "corporation", Fileds: []claFields{{Language: "en"}}},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConfigValidateInvalidCLAType(t *testing.T) {
	cfg := &Config{
		CLA: []claConfig{
			{Type: "invalid_type"},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid CLA type")
	}
}

func TestConfigGetLanguage(t *testing.T) {
	cfg := &Config{
		CLA: []claConfig{
			{Type: "corporation", Fileds: []claFields{
				{Language: "en"},
				{Language: "zh"},
			}},
		},
	}
	cfg.Validate()

	if v := cfg.getLanguage("en"); v != "en" {
		t.Errorf("expected 'en', got '%s'", v)
	}
	if v := cfg.getLanguage("ZH"); v != "zh" {
		t.Errorf("expected 'zh', got '%s'", v)
	}
	if v := cfg.getLanguage("fr"); v != "" {
		t.Errorf("expected '', got '%s'", v)
	}
}

func TestGetCLAFileds(t *testing.T) {
	cfg := &Config{
		CLA: []claConfig{
			{
				Type: "corporation",
				Fileds: []claFields{
					{
						Language: "en",
						Fileds:   []CLAField{{Type: "text", Desc: "Name", Title: "Name", MaxLength: 50}},
					},
				},
			},
		},
	}
	cfg.Validate()
	Init(cfg)

	fields := GetCLAFileds(CLATypeCorp, CreateLanguage("en"))
	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}
	if fields[0].Title != "Name" {
		t.Errorf("expected 'Name', got '%s'", fields[0].Title)
	}
}

func TestGetCLAFiledsNotFound(t *testing.T) {
	fields := GetCLAFileds(CLATypeIndividual, CreateLanguage("zh"))
	if len(fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(fields))
	}
}

// --- CLAField ---

func TestCLAFieldIsValidValue(t *testing.T) {
	f := &CLAField{MaxLength: 10}
	if !f.IsValidValue("short") {
		t.Error("expected valid")
	}
	if f.IsValidValue("this is too long") {
		t.Error("expected invalid for too long value")
	}
	if f.IsValidValue("<script>") {
		t.Error("expected invalid for XSS content")
	}
}
