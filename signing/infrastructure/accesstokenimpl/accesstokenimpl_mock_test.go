package accesstokenimpl

import (
	"errors"
	"testing"
	"time"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

type mockAccessTokenDao struct {
	setErr       error
	getVal       interface{}
	getErr       error
	expireErr    error
	isNotExists  bool
}

func (m *mockAccessTokenDao) Set(key string, val interface{}) error     { return m.setErr }
func (m *mockAccessTokenDao) Get(key string, val interface{}) error     { return m.getErr }
func (m *mockAccessTokenDao) Expire(key string, d time.Duration) error  { return m.expireErr }
func (m *mockAccessTokenDao) IsDocNotExists(err error) bool              { return m.isNotExists }

func TestAccessTokenImplAdd(t *testing.T) {
	dao := &mockAccessTokenDao{}
	impl := &accessTokenImpl{dao: dao, expiry: time.Second}

	at := &domain.AccessToken{
		Expiry:        1000,
		Payload:       []byte("payload"),
		EncryptedCSRF: []byte("csrf"),
	}
	key, err := impl.Add(at)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == "" {
		t.Error("expected non-empty key")
	}
}

func TestAccessTokenImplFind(t *testing.T) {
	// We can't easily test Find without real serialization in redis,
	// but we can test the error paths
	dao := &mockAccessTokenDao{getErr: errors.New("not found"), isNotExists: true}
	impl := &accessTokenImpl{dao: dao}
	_, err := impl.Find("key")
	if err == nil {
		t.Error("expected error")
	}
}

func TestAccessTokenImplDelete(t *testing.T) {
	dao := &mockAccessTokenDao{}
	impl := &accessTokenImpl{dao: dao, expiry: time.Second}
	if err := impl.Delete("key"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewAccessTokenImpl(t *testing.T) {
	cfg := &Config{Expire: 120}
	impl := NewAccessTokenImpl(&mockAccessTokenDao{}, cfg)
	if impl == nil {
		t.Fatal("expected non-nil")
	}
	if impl.expiry != 120*time.Second {
		t.Errorf("expected 120s, got %v", impl.expiry)
	}
}
