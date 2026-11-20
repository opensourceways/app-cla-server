package vcratelimitimpl

import (
	"errors"
	"testing"
	"time"
)

type mockDao struct {
	counts          map[string]int64
	incrErr         error
	expiry          time.Duration
	expireErr       error
	expired         bool
	expireCallCount int
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
	m.expireCallCount++
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

func TestAllowRefreshesExpiryOnSubsequentRequests(t *testing.T) {
	d := newMockDao()
	l := NewVCRateLimiter(d, &Config{Threshold: 5, Window: 30})

	// First request sets expiry
	_, _ = l.Allow("192.168.1.1")
	if d.expireCallCount != 1 {
		t.Fatalf("expected 1 Expire call after first request, got %d", d.expireCallCount)
	}

	// Second request within quota should also refresh expiry
	_, _ = l.Allow("192.168.1.1")
	if d.expireCallCount != 2 {
		t.Fatalf("expected 2 Expire calls after second request, got %d", d.expireCallCount)
	}
	if !d.expired {
		t.Error("expected Expire to be called on subsequent requests within quota")
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

// TestAllowExpireFailureSelfHeals is a regression test for the bug where
// INCR+EXPIRE non-atomicity could leave a key with no TTL, permanently
// banning the IP. With the sliding-window fix, a failed Expire on the
// first request is retried on subsequent requests within the quota,
// self-healing the missing TTL.
func TestAllowExpireFailureSelfHeals(t *testing.T) {
	d := newMockDao()
	l := NewVCRateLimiter(d, &Config{Threshold: 5, Window: 60})

	// First request: Incr succeeds (count=1), Expire fails.
	d.expireErr = errors.New("redis down")
	allowed, err := l.Allow("1.2.3.4")
	if err == nil {
		t.Fatal("expected error when Expire fails on first request")
	}
	if allowed {
		t.Error("expected not allowed when Expire fails")
	}
	if d.expireCallCount != 1 {
		t.Fatalf("expected 1 Expire call, got %d", d.expireCallCount)
	}

	// Second request: Expire now succeeds, self-healing the missing TTL.
	d.expireErr = nil
	allowed, err = l.Allow("1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error on second request: %v", err)
	}
	if !allowed {
		t.Error("expected allowed on second request within quota")
	}
	if d.expireCallCount != 2 {
		t.Fatalf("expected 2 Expire calls, got %d", d.expireCallCount)
	}
	if !d.expired {
		t.Error("expected Expire to succeed on second request, self-healing the TTL")
	}
}

// TestAllowDoesNotSetExpiryWhenOverThreshold verifies that Expire is not
// called once the counter exceeds the threshold. The key already has a TTL
// from previous successful Expire calls and will expire normally.
func TestAllowDoesNotSetExpiryWhenOverThreshold(t *testing.T) {
	d := newMockDao()
	l := NewVCRateLimiter(d, &Config{Threshold: 2, Window: 60})

	// Exhaust the quota (2 requests within threshold).
	for i := 0; i < 2; i++ {
		_, err := l.Allow("10.0.0.1")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
	}
	if d.expireCallCount != 2 {
		t.Fatalf("expected 2 Expire calls within quota, got %d", d.expireCallCount)
	}

	// Third request exceeds threshold; Expire should NOT be called.
	allowed, err := l.Allow("10.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected denied when over threshold")
	}
	if d.expireCallCount != 2 {
		t.Fatalf("expected Expire not to be called when over threshold, got %d calls", d.expireCallCount)
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
