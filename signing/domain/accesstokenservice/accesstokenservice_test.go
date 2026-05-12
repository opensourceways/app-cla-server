package accesstokenservice

import (
	"encoding/base64"
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/encryption"
	"github.com/opensourceways/app-cla-server/signing/domain/randombytes"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

type mockAccessTokenRepo struct {
	addData  *domain.AccessToken
	addIdx   string
	addErr   error
	findData domain.AccessToken
	findErr  error
	delErr   error
}

func (m *mockAccessTokenRepo) Add(v *domain.AccessToken) (string, error) {
	m.addData = v
	return m.addIdx, m.addErr
}
func (m *mockAccessTokenRepo) Find(key string) (domain.AccessToken, error) {
	return m.findData, m.findErr
}
func (m *mockAccessTokenRepo) Delete(key string) error { return m.delErr }

type mockEncryption struct {
	encryptData []byte
	encryptErr  error
	isSameVal   bool
}

func (m *mockEncryption) Encrypt(v []byte) ([]byte, error) { return m.encryptData, m.encryptErr }
func (m *mockEncryption) IsSame(plain, encrypted []byte) bool {
	return m.isSameVal
}

type mockRandomBytes struct {
	data []byte
	err  error
}

func (m *mockRandomBytes) New(n int) ([]byte, error) { return m.data, m.err }

func TestNewAccessTokenService(t *testing.T) {
	s := NewAccessTokenService(&mockAccessTokenRepo{}, &mockEncryption{}, &mockRandomBytes{})
	if s == nil {
		t.Fatal("expected non-nil")
	}
}

func TestAccessTokenServiceAdd(t *testing.T) {
	randomData := make([]byte, csrfTokenLen)
	for i := range randomData {
		randomData[i] = byte(i)
	}
	encrypted := []byte("encrypted-csrf")

	repo := &mockAccessTokenRepo{addIdx: "token-1"}
	encrypt := &mockEncryption{encryptData: encrypted}
	randBytes := &mockRandomBytes{data: randomData}

	s := NewAccessTokenService(repo, encrypt, randBytes)

	k, err := s.Add([]byte("payload-data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k.Id != "token-1" {
		t.Errorf("expected 'token-1', got '%s'", k.Id)
	}
	expectedCSRF := base64.StdEncoding.EncodeToString(randomData)
	if k.CSRF != expectedCSRF {
		t.Errorf("expected '%s', got '%s'", expectedCSRF, k.CSRF)
	}
	if string(repo.addData.Payload) != "payload-data" {
		t.Errorf("expected 'payload-data', got '%s'", string(repo.addData.Payload))
	}
}

func TestAccessTokenServiceValidate(t *testing.T) {
	randomData := make([]byte, csrfTokenLen)
	for i := range randomData {
		randomData[i] = byte(i)
	}
	csrf := base64.StdEncoding.EncodeToString(randomData)

	repo := &mockAccessTokenRepo{
		findData: domain.AccessToken{
			Expiry:        9999999999,
			Payload:       []byte("payload"),
			EncryptedCSRF: []byte("encrypted"),
		},
	}
	encrypt := &mockEncryption{isSameVal: true}
	randBytes := &mockRandomBytes{data: randomData}
	encrypt2 := &mockEncryption{isSameVal: true, encryptData: []byte("enc2")}
	validToken := domain.AccessToken{
		Expiry:        9999999999,
		Payload:       []byte("payload"),
		EncryptedCSRF: []byte("encrypted"),
	}
	repo2 := &mockAccessTokenRepo{
		addIdx:   "new-token",
		findData: validToken,
	}

	// Validate the original token
	s := &accessTokenService{repo: repo, encrypt: encrypt, randomBytes: randBytes}
	payload, err := s.validate(domain.AccessTokenKey{Id: "old", CSRF: csrf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(payload) != "payload" {
		t.Errorf("expected 'payload', got '%s'", string(payload))
	}

	// ValidateAndRefresh
	s2 := &accessTokenService{repo: repo2, encrypt: encrypt2, randomBytes: randBytes}
	newKey, p, err := s2.ValidateAndRefresh(domain.AccessTokenKey{Id: "old", CSRF: csrf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newKey.Id != "new-token" {
		t.Errorf("expected 'new-token', got '%s'", newKey.Id)
	}
	_ = p
}

func TestAccessTokenServiceValidateInvalidCSRF(t *testing.T) {
	s := &accessTokenService{
		repo:    &mockAccessTokenRepo{},
		encrypt: &mockEncryption{},
	}
	_, err := s.validate(domain.AccessTokenKey{CSRF: "invalid-base64!!!"})
	if err == nil {
		t.Error("expected error for invalid base64 CSRF")
	}
}

func TestAccessTokenServiceValidateWrongLength(t *testing.T) {
	short := base64.StdEncoding.EncodeToString([]byte("short"))
	s := &accessTokenService{
		repo:    &mockAccessTokenRepo{},
		encrypt: &mockEncryption{},
	}
	_, err := s.validate(domain.AccessTokenKey{CSRF: short})
	if err == nil {
		t.Error("expected error for short CSRF")
	}
}

func TestAccessTokenServiceValidateNotFound(t *testing.T) {
	randomData := make([]byte, csrfTokenLen)
	csrf := base64.StdEncoding.EncodeToString(randomData)
	repo := &mockAccessTokenRepo{
		findErr: commonRepo.NewErrorResourceNotFound(nil),
	}
	s := &accessTokenService{repo: repo, encrypt: &mockEncryption{}}
	_, err := s.validate(domain.AccessTokenKey{CSRF: csrf})
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestAccessTokenServiceValidateExpired(t *testing.T) {
	randomData := make([]byte, csrfTokenLen)
	csrf := base64.StdEncoding.EncodeToString(randomData)
	repo := &mockAccessTokenRepo{
		findData: domain.AccessToken{
			Expiry:        1, // expired
			Payload:       []byte("p"),
			EncryptedCSRF: []byte("c"),
		},
	}
	s := &accessTokenService{repo: repo, encrypt: &mockEncryption{}}
	_, err := s.validate(domain.AccessTokenKey{CSRF: csrf})
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestAccessTokenServiceValidateWrongCSRF(t *testing.T) {
	randomData := make([]byte, csrfTokenLen)
	csrf := base64.StdEncoding.EncodeToString(randomData)
	repo := &mockAccessTokenRepo{
		findData: domain.AccessToken{
			Expiry:        9999999999,
			Payload:       []byte("p"),
			EncryptedCSRF: []byte("real-encrypted"),
		},
	}
	encrypt := &mockEncryption{isSameVal: false} // CSRF doesn't match
	s := &accessTokenService{repo: repo, encrypt: encrypt}
	_, err := s.validate(domain.AccessTokenKey{CSRF: csrf})
	if err == nil {
		t.Error("expected error for wrong CSRF")
	}
}

// Verify interface compliance
var _ AccessTokenService = (*accessTokenService)(nil)
var _ repository.AccessToken = (*mockAccessTokenRepo)(nil)
var _ encryption.Encryption = (*mockEncryption)(nil)
var _ randombytes.RandomBytes = (*mockRandomBytes)(nil)
