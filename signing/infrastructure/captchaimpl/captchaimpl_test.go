package captchaimpl

import (
	"testing"
	"time"
)

type mockDao struct {
	data map[string]string
}

func newMockDao() *mockDao {
	return &mockDao{data: make(map[string]string)}
}

func (m *mockDao) SetWithExpiry(key string, val interface{}, expiry time.Duration) error {
	m.data[key] = val.(string)
	return nil
}

func (m *mockDao) Get(key string, val interface{}) error {
	v, ok := m.data[key]
	if !ok {
		return errNotFound
	}
	*(val.(*string)) = v
	return nil
}

func (m *mockDao) Del(keys ...string) error {
	for _, k := range keys {
		delete(m.data, k)
	}
	return nil
}

func (m *mockDao) IsDocNotExists(err error) bool {
	return err == errNotFound
}

var errNotFound = &notFoundErr{}

type notFoundErr struct{}

func (e *notFoundErr) Error() string { return "not found" }

func TestCaptchaStoreSetGet(t *testing.T) {
	d := newMockDao()
	store := &captchaStore{dao: d, expiry: 5 * time.Minute}

	if err := store.Set("abc", "1234"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val := store.Get("abc", false)
	if val != "1234" {
		t.Errorf("expected '1234', got '%s'", val)
	}
}

func TestCaptchaStoreGetClear(t *testing.T) {
	d := newMockDao()
	store := &captchaStore{dao: d, expiry: 5 * time.Minute}

	_ = store.Set("abc", "1234")
	val := store.Get("abc", true)
	if val != "1234" {
		t.Errorf("expected '1234', got '%s'", val)
	}

	val2 := store.Get("abc", false)
	if val2 != "" {
		t.Errorf("expected empty after clear, got '%s'", val2)
	}
}

func TestCaptchaStoreVerify(t *testing.T) {
	d := newMockDao()
	store := &captchaStore{dao: d, expiry: 5 * time.Minute}

	_ = store.Set("abc", "1234")
	if !store.Verify("abc", "1234", true) {
		t.Error("expected verify to pass")
	}

	_ = store.Set("def", "5678")
	if store.Verify("def", "wrong", true) {
		t.Error("expected verify to fail")
	}
}

func TestCaptchaStoreVerifyCaseInsensitive(t *testing.T) {
	d := newMockDao()
	store := &captchaStore{dao: d, expiry: 5 * time.Minute}

	_ = store.Set("abc", "ABCD")
	if !store.Verify("abc", "abcd", true) {
		t.Error("expected case-insensitive verify to pass")
	}
}

func TestCaptchaStoreGetMissing(t *testing.T) {
	d := newMockDao()
	store := &captchaStore{dao: d, expiry: 5 * time.Minute}

	val := store.Get("nonexistent", false)
	if val != "" {
		t.Errorf("expected empty for missing key, got '%s'", val)
	}
}

func TestCaptchaKey(t *testing.T) {
	if key := captchaKey("abc"); key != "captcha:abc" {
		t.Errorf("expected 'captcha:abc', got '%s'", key)
	}
}
