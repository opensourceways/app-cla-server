package emailcredential

import (
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/symmetricencryption"
)

type mockEmailCredentialRepo struct {
	addData *domain.EmailCredential
	addErr  error
	findRes domain.EmailCredential
	findErr error
}

func (m *mockEmailCredentialRepo) Add(e *domain.EmailCredential) error { m.addData = e; return m.addErr }
func (m *mockEmailCredentialRepo) Find(ed dp.EmailAddr) (domain.EmailCredential, error) {
	return m.findRes, m.findErr
}

type mockSymEncrypt2 struct {
	encryptRes []byte
	encryptErr error
	decryptRes []byte
	decryptErr error
}

func (m *mockSymEncrypt2) Encrypt(plain []byte) ([]byte, error) { return m.encryptRes, m.encryptErr }
func (m *mockSymEncrypt2) Decrypt(cipher []byte) ([]byte, error) { return m.decryptRes, m.decryptErr }

func TestNewEmailCredential(t *testing.T) {
	s := NewEmailCredential(&mockEmailCredentialRepo{}, &mockSymEncrypt2{})
	if s == nil {
		t.Fatal("expected non-nil")
	}
}

func TestEmailCredentialAdd(t *testing.T) {
	encrypt := &mockSymEncrypt2{encryptRes: []byte("encrypted-token")}
	repo := &mockEmailCredentialRepo{}
	s := NewEmailCredential(repo, encrypt)

	e := &domain.EmailCredential{
		Addr:     dp.CreateEmailAddr("test@example.com"),
		Token:    []byte("raw-token"),
		Platform: "gmail",
	}
	if err := s.Add(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(repo.addData.Token) != "encrypted-token" {
		t.Errorf("expected 'encrypted-token', got '%s'", string(repo.addData.Token))
	}
}

func TestEmailCredentialFind(t *testing.T) {
	encryptedToken := []byte("encrypted-token")
	decryptedToken := []byte("decrypted-token")

	repo := &mockEmailCredentialRepo{
		findRes: domain.EmailCredential{
			Addr:     dp.CreateEmailAddr("test@example.com"),
			Token:    encryptedToken,
			Platform: "gmail",
		},
	}
	encrypt := &mockSymEncrypt2{decryptRes: decryptedToken}
	s := NewEmailCredential(repo, encrypt)

	e, err := s.Find(dp.CreateEmailAddr("test@example.com"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(e.Token) != "decrypted-token" {
		t.Errorf("expected 'decrypted-token', got '%s'", string(e.Token))
	}
}

func TestEmailCredentialFindNotFound(t *testing.T) {
	repo := &mockEmailCredentialRepo{
		findErr: commonRepo.NewErrorResourceNotFound(nil),
	}
	s := NewEmailCredential(repo, &mockSymEncrypt2{})

	_, err := s.Find(dp.CreateEmailAddr("missing@test.com"))
	if err == nil {
		t.Error("expected error for not found")
	}
}

// Verify interface compliance
var _ EmailCredential = (*emailCredential)(nil)
var _ repository.EmailCredential = (*mockEmailCredentialRepo)(nil)
var _ symmetricencryption.Encryption = (*mockSymEncrypt2)(nil)
