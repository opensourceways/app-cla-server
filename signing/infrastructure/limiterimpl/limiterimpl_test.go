package limiterimpl

import (
	"testing"
	"time"
)

type mockLimiterDao struct {
	setKeyErr   error
	hasKeyRes   bool
	hasKeyErr   error
}

func (m *mockLimiterDao) SetKey(k string, expiry time.Duration) error { return m.setKeyErr }
func (m *mockLimiterDao) HasKey(k string) (bool, error)               { return m.hasKeyRes, m.hasKeyErr }

func TestNewLimiterImpl(t *testing.T) {
	impl := NewLimiterImpl(&mockLimiterDao{})
	if impl == nil {
		t.Fatal("expected non-nil")
	}
}

func TestLimiterAdd(t *testing.T) {
	impl := &limiterImpl{dao: &mockLimiterDao{}}
	if err := impl.Add("key", time.Second); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLimiterIsAllowed(t *testing.T) {
	// Key doesn't exist → allowed
	impl := &limiterImpl{dao: &mockLimiterDao{hasKeyRes: false}}
	allowed, err := impl.IsAllowed("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected allowed")
	}

	// Key exists → not allowed
	impl2 := &limiterImpl{dao: &mockLimiterDao{hasKeyRes: true}}
	allowed2, _ := impl2.IsAllowed("key")
	if allowed2 {
		t.Error("expected not allowed")
	}
}
