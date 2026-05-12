package loginimpl

import (
	"errors"
	"testing"
	"time"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

type mockLoginDao struct {
	setErr      error
	getErr      error
	expireErr   error
	isNotExists bool
}

func (m *mockLoginDao) SetWithExpiry(key string, val interface{}, expiry time.Duration) error {
	return m.setErr
}
func (m *mockLoginDao) Get(key string, val interface{}) error     { return m.getErr }
func (m *mockLoginDao) Expire(key string, d time.Duration) error  { return m.expireErr }
func (m *mockLoginDao) IsDocNotExists(err error) bool              { return m.isNotExists }

func TestLoginImplAdd(t *testing.T) {
	dao := &mockLoginDao{}
	impl := &loginImpl{dao: dao, expiry: time.Minute}

	l := &domain.Login{Id: "test-id", Frozen: false, FailedNum: 2}
	if err := impl.Add(l); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoginImplFind(t *testing.T) {
	dao := &mockLoginDao{getErr: errors.New("not found"), isNotExists: true}
	impl := &loginImpl{dao: dao}
	_, err := impl.Find("key")
	if err == nil {
		t.Error("expected error")
	}
}

func TestLoginImplDelete(t *testing.T) {
	dao := &mockLoginDao{}
	impl := &loginImpl{dao: dao}
	if err := impl.Delete("key"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewLoginImpl(t *testing.T) {
	cfg := &Config{Expire: 600}
	impl := NewLoginImpl(&mockLoginDao{}, cfg)
	if impl == nil {
		t.Fatal("expected non-nil")
	}
	if impl.expiry != 600*time.Second {
		t.Errorf("expected 600s, got %v", impl.expiry)
	}
}
