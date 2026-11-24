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

// 验证窗口内请求数未超阈值时全部放行
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

// 验证超过阈值后请求被拒绝
func TestAllowExceedsThreshold(t *testing.T) {
	d := newMockDao()
	l := NewVCRateLimiter(d, &Config{Threshold: 2, Window: 60})

	for i := 0; i < 2; i++ {
		allowed, err := l.Allow("10.0.0.1")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
		if !allowed {
			t.Fatalf("request %d: expected allowed, got denied", i)
		}
	}

	allowed, err := l.Allow("10.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected denied after exceeding threshold, got allowed")
	}
}

// 验证仅在窗口第一次请求时设置过期时间
func TestAllowSetsExpiryOnlyOnFirstRequest(t *testing.T) {
	d := newMockDao()
	l := NewVCRateLimiter(d, &Config{Threshold: 5, Window: 30})

	for i := 0; i < 3; i++ {
		_, err := l.Allow("192.168.1.1")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
	}

	if !d.expired {
		t.Error("expected Expire to be called")
	}
	if d.expiry != 30*time.Second {
		t.Errorf("expected expiry of 30s, got %v", d.expiry)
	}
	if d.expireCallCount != 1 {
		t.Errorf("expected 1 Expire call (first request only), got %d", d.expireCallCount)
	}
}

// 验证Incr失败时返回错误且拒绝请求
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

// 验证第一次请求设置过期时间失败时返回错误
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

// 验证不同IP的计数互不影响
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
