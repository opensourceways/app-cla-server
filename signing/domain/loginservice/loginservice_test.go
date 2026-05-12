package loginservice

import (
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/encryption"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/userpassword"
)

func init() {
	// Setup domain config with login settings
	domain.Init(&domain.Config{
		MaxNumOfFailedLogin:  5,
		NeedCaptchaThreshold: 1,
	})
}

type mockUserRepo struct {
	findByAccountRes domain.User
	findByAccountErr error
	findByEmailRes   domain.User
	findByEmailErr   error
}

func (m *mockUserRepo) Find(string) (domain.User, error)                      { return domain.User{}, nil }
func (m *mockUserRepo) FindByAccount(linkId string, a dp.Account) (domain.User, error) {
	return m.findByAccountRes, m.findByAccountErr
}
func (m *mockUserRepo) FindByEmail(linkId string, e dp.EmailAddr) (domain.User, error) {
	return m.findByEmailRes, m.findByEmailErr
}
func (m *mockUserRepo) FindAllByLinkId(string) ([]domain.User, error)          { return nil, nil }
func (m *mockUserRepo) Add(*domain.User) (string, error)                       { return "", nil }
func (m *mockUserRepo) AddForMigrate(*domain.User) (string, error)             { return "", nil }
func (m *mockUserRepo) Remove([]string) error                                  { return nil }
func (m *mockUserRepo) RemoveByAccount(string, []dp.Account) error            { return nil }
func (m *mockUserRepo) SavePassword(*domain.User) error                        { return nil }
func (m *mockUserRepo) SavePrivacyConsent(*domain.User) error                  { return nil }

type mockLoginRepo struct {
	findRes domain.Login
	findErr error
	addErr  error
	delErr  error
}

func (m *mockLoginRepo) Add(l *domain.Login) error { return m.addErr }
func (m *mockLoginRepo) Find(key string) (domain.Login, error) {
	return m.findRes, m.findErr
}
func (m *mockLoginRepo) Delete(key string) error { return m.delErr }

type mockEncrypt struct {
	isSameVal   bool
	encryptRes  []byte
	encryptErr  error
}

func (m *mockEncrypt) Encrypt(v []byte) ([]byte, error) { return m.encryptRes, m.encryptErr }
func (m *mockEncrypt) IsSame(plain, encrypted []byte) bool {
	return m.isSameVal
}

type mockPassword struct {
	valid bool
}

func (m *mockPassword) New() (dp.Password, error) { return dp.NewPassword([]byte("new")) }
func (m *mockPassword) IsValid(dp.Password) bool   { return m.valid }

func TestNewLoginService(t *testing.T) {
	s := NewLoginService(&mockUserRepo{}, &mockLoginRepo{}, &mockEncrypt{}, &mockPassword{})
	if s == nil {
		t.Fatal("expected non-nil")
	}
}

func TestLoginByAccountSuccess(t *testing.T) {
	pw, _ := dp.NewPassword([]byte("correct"))
	encPassword := []byte("encrypted-password")

	userRepo := &mockUserRepo{
		findByAccountRes: domain.User{
			UserBasicInfo: domain.UserBasicInfo{Password: encPassword},
		},
	}
	loginRepo := &mockLoginRepo{
		findErr: commonRepo.NewErrorResourceNotFound(nil),
	}
	encrypt := &mockEncrypt{isSameVal: true}
	password := &mockPassword{valid: true}

	s := NewLoginService(userRepo, loginRepo, encrypt, password)
	u, lv, err := s.LoginByAccount("link1", dp.CreateAccount("user"), pw)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	_ = u
	_ = lv
}

func TestLoginByAccountWrongPassword(t *testing.T) {
	pw, _ := dp.NewPassword([]byte("wrong"))
	encPassword := []byte("encrypted-password")

	userRepo := &mockUserRepo{
		findByAccountRes: domain.User{
			UserBasicInfo: domain.UserBasicInfo{Password: encPassword},
		},
	}
	loginRepo := &mockLoginRepo{
		findErr: commonRepo.NewErrorResourceNotFound(nil),
	}
	encrypt := &mockEncrypt{isSameVal: false} // wrong password
	password := &mockPassword{valid: true}

	s := NewLoginService(userRepo, loginRepo, encrypt, password)
	_, _, err := s.LoginByAccount("link1", dp.CreateAccount("user"), pw)
	if err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestLoginByAccountFrozen(t *testing.T) {
	pw, _ := dp.NewPassword([]byte("correct"))

	loginRepo := &mockLoginRepo{
		findRes: domain.Login{Frozen: true},
	}
	password := &mockPassword{valid: true}

	s := NewLoginService(&mockUserRepo{}, loginRepo, &mockEncrypt{}, password)
	_, _, err := s.LoginByAccount("link1", dp.CreateAccount("user"), pw)
	if err == nil {
		t.Error("expected error for frozen login")
	}
}

func TestNeedCaptcha(t *testing.T) {
	loginRepo := &mockLoginRepo{
		findRes: domain.Login{FailedNum: 2}, // threshold is 1, so NeedCaptcha = true
	}
	s := NewLoginService(nil, loginRepo, nil, nil)
	b, err := s.NeedCaptcha("test-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !b {
		t.Error("expected captcha needed")
	}
}

func TestNeedCaptchaNotFound(t *testing.T) {
	loginRepo := &mockLoginRepo{
		findErr: commonRepo.NewErrorResourceNotFound(nil),
	}
	s := NewLoginService(nil, loginRepo, nil, nil)
	b, err := s.NeedCaptcha("test-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b {
		t.Error("expected no captcha for not found")
	}
}

func TestClearLoginFailure(t *testing.T) {
	loginRepo := &mockLoginRepo{}
	s := NewLoginService(nil, loginRepo, nil, nil)
	if err := s.ClearLoginFailure("test-id"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetLoginInfo(t *testing.T) {
	loginRepo := &mockLoginRepo{
		findRes: domain.Login{Id: "test", FailedNum: 2},
	}
	s := NewLoginService(nil, loginRepo, nil, nil)
	lv, err := s.GetLoginInfo("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lv == nil || lv.FailedNum != 2 {
		t.Errorf("expected FailedNum=2, got %+v", lv)
	}
}

func TestGetLoginInfoNotFound(t *testing.T) {
	loginRepo := &mockLoginRepo{
		findErr: commonRepo.NewErrorResourceNotFound(nil),
	}
	s := NewLoginService(nil, loginRepo, nil, nil)
	lv, err := s.GetLoginInfo("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lv != nil {
		t.Error("expected nil for not found")
	}
}

// Verify interface compliance
var _ LoginService = (*loginService)(nil)
var _ repository.User = (*mockUserRepo)(nil)
var _ repository.Login = (*mockLoginRepo)(nil)
var _ encryption.Encryption = (*mockEncrypt)(nil)
var _ userpassword.UserPassword = (*mockPassword)(nil)
