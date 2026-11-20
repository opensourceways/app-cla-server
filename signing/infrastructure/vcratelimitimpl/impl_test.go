package vcratelimitimpl

import (
	"errors"
	"testing"
	"time"
)

type mockDao struct {
	counts   map[string]int64
	incrErr  error
	expiry   time.Duration
	expireErr error
	expired  bool
}

func newMockDao() *mockDao {
	return &mockDao{counts: make(map[string]int64)}
}

func (m *mockDao) Incr(key string) (int64, error) {
	if m.incrErr != nil {
		return 0, m.incrErr
	}
	m.counts[key]++
	return m.counts[key], nil
}

func (m *mockDao) Expire(key string, expire time.Duration) error {
	if m.expireErr != nil {
		return m.expireErr
	}
	m.expiry = expire
	m.expired = true
	return nil
}

func TestAllowUnderThreshold(t *testing.T) {
	d := newMockDao()
	l := NewVCRateLimiter(d, &Config{Threshold: 3, Window: 60})

	for i := int64(1); i <= 3; i++ {
		allowed, err := l.Allow("1.2.3.4")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
		if !allowed {
			t.Errorf("request %d: expected allowed, got denied", i)
		}
	}
}

func TestAllowExceedsThreshold(t *testing.T) {
	d := newMockDao()
	l := NewVCRateLimiter(d, &Config{Threshold: 2, Window: 60})

	// First two requests are allowed
	for i := 0; i < 2; i++ {
		allowed, err := l.Allow("10.0.0.1")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
		if !allowed {
			t.Fatalf("request %d: expected allowed, got denied", i)
		}
	}

	// Third request should be denied
	allowed, err := l.Allow("10.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected denied after exceeding threshold, got allowed")
	}
}

func TestAllowSetsExpiryOnFirstRequest(t *testing.T) {
	d := newMockDao()
	l := NewVCRateLimiter(d, &Config{Threshold: 5, Window: 30})

	allowed, err := l.Allow("192.168.1.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected allowed on first request")
	}

	if !d.expired {
		t.Error("expected Expire to be called on first request")
	}
	if d.expiry != 30*time.Second {
		t.Errorf("expected expiry of 30s, got %v", d.expiry)
	}
}

func TestAllowDoesNotSetExpiryOnSubsequentRequests(t *testing.T) {
	d := newMockDao()
	l := NewVCRateLimiter(d, &Config{Threshold: 5, Window: 30})

	// First request sets expiry
	_, _ = l.Allow("192.168.1.1")
	d.expired = false // reset

	// Second request should not set expiry
	_, _ = l.Allow("192.168.1.1")
	if d.expired {
		t.Error("expected Expire not to be called on subsequent requests")
	}
}

func TestAllowIncrError(t *testing.T) {
	d := newMockDao()
	d.incrErr = errors.New("redis down")
	l := NewVCRateLimiter(d, &Config{Threshold: 5, Window: 60})

	allowed, err := l.Allow("1.2.3.4")
	if err == nil {
		t.Fatal("expected error when Incr fails")
	}
	if allowed {
		t.Error("expected not allowed when Incr fails")
	}
}

func TestAllowExpireErrorOnFirstRequest(t *testing.T) {
	d := newMockDao()
	d.expireErr = errors.New("redis down")
	l := NewVCRateLimiter(d, &Config{Threshold: 5, Window: 60})

	allowed, err := l.Allow("1.2.3.4")
	if err == nil {
		t.Fatal("expected error when Expire fails on first request")
	}
	if allowed {
		t.Error("expected not allowed when Expire fails")
	}
}

func TestAllowDifferentIPsIndependent(t *testing.T) {
	d := newMockDao()
	l := NewVCRateLimiter(d, &Config{Threshold: 1, Window: 60})

	allowed1, err := l.Allow("1.1.1.1")
	if err != nil || !allowed1 {
		t.Fatalf("first IP should be allowed, got allowed=%v err=%v", allowed1, err)
	}

	allowed2, err := l.Allow("2.2.2.2")
	if err != nil || !allowed2 {
		t.Fatalf("second IP should be allowed independently, got allowed=%v err=%v", allowed2, err)
	}
}
