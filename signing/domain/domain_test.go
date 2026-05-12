package domain

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func init() {
	config = Config{
		IsTestEnvironment:       false,
		MaxNumOfFailedLogin:     5,
		NeedCaptchaThreshold:    1,
		AccessTokenExpiry:       3600,
		CommunityManagerLinkId:  "community_link",
		InvalidCorpEmailDomain:  "gmail.com,yahoo.com",
	}
}

// --- Login ---

func TestNewLogin(t *testing.T) {
	l := NewLogin("test-id")
	if l.Id != "test-id" {
		t.Errorf("expected 'test-id', got '%s'", l.Id)
	}
	if l.Frozen {
		t.Error("expected not frozen")
	}
	if l.FailedNum != 0 {
		t.Errorf("expected 0, got %d", l.FailedNum)
	}
}

func TestLoginFail(t *testing.T) {
	l := NewLogin("test")

	// Fail 4 times should not freeze
	for i := 1; i < config.MaxNumOfFailedLogin; i++ {
		if l.Fail() {
			t.Errorf("expected false on fail %d/%d", i, config.MaxNumOfFailedLogin)
		}
		if l.FailedNum != i {
			t.Errorf("expected FailedNum=%d, got %d", i, l.FailedNum)
		}
	}

	// 5th failure should freeze
	if !l.Fail() {
		t.Error("expected true on final failure")
	}
	if l.FailedNum != config.MaxNumOfFailedLogin {
		t.Errorf("expected FailedNum=%d, got %d", config.MaxNumOfFailedLogin, l.FailedNum)
	}
	if !l.Frozen {
		t.Error("expected frozen after max failures")
	}
}

func TestLoginRetryNum(t *testing.T) {
	l := NewLogin("test")
	if n := l.RetryNum(); n != config.MaxNumOfFailedLogin {
		t.Errorf("expected %d, got %d", config.MaxNumOfFailedLogin, n)
	}

	l.Fail()
	if n := l.RetryNum(); n != config.MaxNumOfFailedLogin-1 {
		t.Errorf("expected %d, got %d", config.MaxNumOfFailedLogin-1, n)
	}

	l.Frozen = true
	if n := l.RetryNum(); n != 0 {
		t.Errorf("expected 0 when frozen, got %d", n)
	}
}

func TestLoginHasFailure(t *testing.T) {
	l := NewLogin("test")
	if l.HasFailure() {
		t.Error("expected false for new login")
	}

	l.Fail()
	if !l.HasFailure() {
		t.Error("expected true after failure")
	}
}

func TestLoginNeedCaptcha(t *testing.T) {
	l := NewLogin("test")

	// Fresh login should not need captcha
	if l.NeedCaptcha() {
		t.Error("expected false for new login")
	}

	// After 1 failure (threshold=1), should need captcha
	l.Fail()
	if !l.NeedCaptcha() {
		t.Error("expected true after reaching threshold")
	}

	// Frozen should not need captcha
	l.Frozen = true
	if l.NeedCaptcha() {
		t.Error("expected false when frozen")
	}
}

// --- CLA ---

func TestCLAIsMe(t *testing.T) {
	lang := dp.CreateLanguage("en")
	cla1 := &CLA{Type: dp.CLATypeCorp, Language: lang}
	cla2 := &CLA{Type: dp.CLATypeCorp, Language: lang}
	cla3 := &CLA{Type: dp.CLATypeIndividual, Language: lang}

	if !cla1.isMe(cla2) {
		t.Error("expected same type and language")
	}
	if cla1.isMe(cla3) {
		t.Error("expected different type")
	}
}

// --- Link ---

func TestLinkCanDo(t *testing.T) {
	link := &Link{Submitter: "admin"}
	if err := link.CanDo("admin"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := link.CanDo("other"); err == nil {
		t.Error("expected error for different user")
	}
}

func TestLinkAddCLA(t *testing.T) {
	link := &Link{CLANum: 0}
	cla := &CLA{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")}

	if err := link.AddCLA(cla); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cla.Id != "0" {
		t.Errorf("expected Id '0', got '%s'", cla.Id)
	}
	if link.CLANum != 1 {
		t.Errorf("expected CLANum=1, got %d", link.CLANum)
	}

	// After AddCLA assigns an ID, simulate the CLA being stored in the link's list.
	link.CLAs = append(link.CLAs, *cla)

	// Adding same type+language CLA should fail
	dup := &CLA{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")}
	if err := link.AddCLA(dup); err == nil {
		t.Error("expected error for duplicate CLA")
	}
}

func TestLinkUpdateCLA(t *testing.T) {
	link := &Link{CLANum: 0}
	cla := &CLA{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")}
	link.CLAs = append(link.CLAs, *cla)

	newCLA := &CLA{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")}
	if err := link.UpdateCLA(newCLA); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if newCLA.Id != "0" {
		t.Errorf("expected Id '0', got '%s'", newCLA.Id)
	}

	// Updating non-existent CLA type should fail
	unknown := &CLA{Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")}
	if err := link.UpdateCLA(unknown); err == nil {
		t.Error("expected error for non-existent CLA type")
	}
}

func TestLinkFindCLA(t *testing.T) {
	link := &Link{
		CLAs: []CLA{
			{Id: "0", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
			{Id: "1", Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")},
		},
	}

	if c := link.FindCLA("0"); c == nil || c.Id != "0" {
		t.Error("expected to find CLA with Id '0'")
	}
	if c := link.FindCLA("2"); c != nil {
		t.Error("expected nil for non-existent CLA")
	}
}

func TestLinkGetCLA(t *testing.T) {
	link := &Link{
		CLAs: []CLA{
			{Id: "0", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
			{Id: "1", Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")},
		},
	}

	c := link.GetCLA(dp.CLATypeCorp, dp.CreateLanguage("en"))
	if c == nil || c.Id != "0" {
		t.Error("expected CLA with Id '0'")
	}

	c = link.GetCLA(dp.CLATypeIndividual, dp.CreateLanguage("zh"))
	if c != nil {
		t.Error("expected nil for missing language")
	}
}

// --- User / UserBasicInfo ---

func TestUserIsCommunityManager(t *testing.T) {
	u := &User{LinkId: "community_link"}
	if !u.IsCommunityManager() {
		t.Error("expected true for community manager")
	}

	u2 := &User{LinkId: "other_link"}
	if u2.IsCommunityManager() {
		t.Error("expected false for non-community manager")
	}
}

func TestUserBasicInfoResetPassword(t *testing.T) {
	var oldPw = []byte("old")
	u := &UserBasicInfo{Password: oldPw}
	u.ResetPassword([]byte("new"))
	if string(u.Password) != "new" {
		t.Errorf("expected 'new', got '%s'", string(u.Password))
	}
	if !u.PasswordChanged {
		t.Error("expected PasswordChanged=true")
	}
}

func TestUserBasicInfoChangePassword(t *testing.T) {
	u := &UserBasicInfo{Password: []byte("old")}

	isCorrect := func(ciphertext []byte) bool {
		return string(ciphertext) == "old"
	}
	genNew := func() ([]byte, error) {
		return []byte("new"), nil
	}

	if err := u.ChangePassword(isCorrect, genNew); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(u.Password) != "new" {
		t.Errorf("expected 'new', got '%s'", string(u.Password))
	}
}

func TestUserBasicInfoChangePasswordWrongOld(t *testing.T) {
	u := &UserBasicInfo{Password: []byte("old")}

	isCorrect := func(ciphertext []byte) bool {
		return false
	}
	genNew := func() ([]byte, error) {
		return []byte("new"), nil
	}

	if err := u.ChangePassword(isCorrect, genNew); err == nil {
		t.Error("expected error for wrong old password")
	}
}

func TestUserBasicInfoUpdatePrivacyConsent(t *testing.T) {
	u := &UserBasicInfo{}

	if !u.UpdatePrivacyConsent("v1") {
		t.Error("expected true for first consent")
	}
	if u.PrivacyConsent.Version != "v1" {
		t.Errorf("expected 'v1', got '%s'", u.PrivacyConsent.Version)
	}

	if u.UpdatePrivacyConsent("v1") {
		t.Error("expected false for same version")
	}

	if !u.UpdatePrivacyConsent("v2") {
		t.Error("expected true for new version")
	}
	if u.PrivacyConsent.Version != "v2" {
		t.Errorf("expected 'v2', got '%s'", u.PrivacyConsent.Version)
	}
}

// --- AccessToken ---

func TestNewAccessToken(t *testing.T) {
	token := NewAccessToken([]byte("payload"), []byte("csrf"))
	if string(token.Payload) != "payload" {
		t.Errorf("expected 'payload', got '%s'", string(token.Payload))
	}
	if string(token.EncryptedCSRF) != "csrf" {
		t.Errorf("expected 'csrf', got '%s'", string(token.EncryptedCSRF))
	}
}

func TestAccessTokenIsValid(t *testing.T) {
	token := NewAccessToken([]byte("payload"), []byte("csrf"))
	if !token.IsValid() {
		t.Error("expected valid token")
	}
}

func TestAccessTokenIsExpired(t *testing.T) {
	token := &AccessToken{
		Expiry:        100, // very old timestamp
		Payload:       []byte("p"),
		EncryptedCSRF: []byte("c"),
	}
	if token.IsValid() {
		t.Error("expected expired token to be invalid")
	}
}

// --- Config ---

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()

	if cfg.VerificationCodeExpiry != 300 {
		t.Errorf("expected 300, got %d", cfg.VerificationCodeExpiry)
	}
	if cfg.AccessTokenExpiry != 3600 {
		t.Errorf("expected 3600, got %d", cfg.AccessTokenExpiry)
	}
	if cfg.IntervalOfCreatingVC != 60 {
		t.Errorf("expected 60, got %d", cfg.IntervalOfCreatingVC)
	}
	if cfg.MaxNumOfEmployeeManager != 5 {
		t.Errorf("expected 5, got %d", cfg.MaxNumOfEmployeeManager)
	}
	if cfg.MaxSizeOfCLAContent <= 0 {
		t.Error("expected positive default for max size")
	}
	if cfg.MaxNumOfFailedLogin != 5 {
		t.Errorf("expected 5, got %d", cfg.MaxNumOfFailedLogin)
	}
	if cfg.NeedCaptchaThreshold != 1 {
		t.Errorf("expected 1, got %d", cfg.NeedCaptchaThreshold)
	}
	if len(cfg.SourceOfCLAPDF) == 0 {
		t.Error("expected default sources")
	}
	if cfg.FileTypeOfCLAContent != "pdf" {
		t.Errorf("expected 'pdf', got '%s'", cfg.FileTypeOfCLAContent)
	}
	if cfg.CommunityManagerLinkId != "fake_link" {
		t.Errorf("expected 'fake_link', got '%s'", cfg.CommunityManagerLinkId)
	}
	if cfg.GetIntervalOfCreatingVC() <= 0 {
		t.Error("expected positive interval")
	}
}

func TestConfigSetDefaultKeepsCustomValues(t *testing.T) {
	cfg := &Config{
		VerificationCodeExpiry:  600,
		AccessTokenExpiry:       7200,
		MaxNumOfEmployeeManager: 10,
		MaxNumOfFailedLogin:     3,
		NeedCaptchaThreshold:    2,
		FileTypeOfCLAContent:    "docx",
	}
	cfg.SetDefault()

	if cfg.VerificationCodeExpiry != 600 {
		t.Errorf("expected 600, got %d", cfg.VerificationCodeExpiry)
	}
	if cfg.AccessTokenExpiry != 7200 {
		t.Errorf("expected 7200, got %d", cfg.AccessTokenExpiry)
	}
	if cfg.MaxNumOfEmployeeManager != 10 {
		t.Errorf("expected 10, got %d", cfg.MaxNumOfEmployeeManager)
	}
}

func TestConfigInvalidCorpEmailDomains(t *testing.T) {
	cfg := &Config{InvalidCorpEmailDomain: "gmail.com,yahoo.com"}
	domains := cfg.InvalidCorpEmailDomains()
	if len(domains) != 2 {
		t.Fatalf("expected 2, got %d", len(domains))
	}
	if domains[0] != "gmail.com" {
		t.Errorf("expected 'gmail.com', got '%s'", domains[0])
	}
	if domains[1] != "yahoo.com" {
		t.Errorf("expected 'yahoo.com', got '%s'", domains[1])
	}
}

func TestConfigInvalidCorpEmailDomainsEmpty(t *testing.T) {
	cfg := &Config{}
	domains := cfg.InvalidCorpEmailDomains()
	if len(domains) != 1 || domains[0] != "" {
		t.Errorf("expected single empty string, got %v", domains)
	}
}

func TestConfigTestEnvironment(t *testing.T) {
	cfg := &Config{IsTestEnvironment: true}
	cfg.setTestDefaults()
	if cfg.TestCaptchaAnswer != "123456" {
		t.Errorf("expected '123456', got '%s'", cfg.TestCaptchaAnswer)
	}

	// Custom test answer should be preserved
	cfg2 := &Config{IsTestEnvironment: true, TestCaptchaAnswer: "999999"}
	cfg2.setTestDefaults()
	if cfg2.TestCaptchaAnswer != "999999" {
		t.Errorf("expected '999999', got '%s'", cfg2.TestCaptchaAnswer)
	}
}
