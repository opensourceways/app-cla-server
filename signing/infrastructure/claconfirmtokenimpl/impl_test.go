package claconfirmtokenimpl

import (
	"encoding"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

var errFakeNotExists = errors.New("fake doc doesn't exist")

// fakeDAO simulates the redis client semantics used by the implementation:
// values are stored as strings, GetAndDelete is atomic (guarded by mutex).
type fakeDAO struct {
	mu       sync.Mutex
	data     map[string]string
	expiry   map[string]time.Duration
	failOn   map[string]error
	delAll   error
	gadCalls int

	// failSetExcept/failSetErr: make SetWithExpiry fail for every key except
	// this one (used to fail the write of randomly generated token keys).
	failSetExcept string
	failSetErr    error
}

func newFakeDAO() *fakeDAO {
	return &fakeDAO{
		data:   map[string]string{},
		expiry: map[string]time.Duration{},
		failOn: map[string]error{},
	}
}

func (d *fakeDAO) SetWithExpiry(key string, val interface{}, expiry time.Duration) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err, ok := d.failOn[key]; ok {
		return err
	}

	if d.failSetErr != nil && key != d.failSetExcept {
		return d.failSetErr
	}

	d.data[key] = marshalValue(val)
	d.expiry[key] = expiry

	return nil
}

func (d *fakeDAO) Get(key string, val interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err, ok := d.failOn[key]; ok {
		return err
	}

	s, ok := d.data[key]
	if !ok {
		return errFakeNotExists
	}

	return scanInto(s, val)
}

func (d *fakeDAO) GetAndDelete(key string, val interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.gadCalls++

	if err, ok := d.failOn[key]; ok {
		return err
	}

	s, ok := d.data[key]
	if !ok {
		return errFakeNotExists
	}

	delete(d.data, key)
	delete(d.expiry, key)

	return scanInto(s, val)
}

func (d *fakeDAO) Del(keys ...string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.delAll != nil {
		return d.delAll
	}

	for _, key := range keys {
		delete(d.data, key)
		delete(d.expiry, key)
	}

	return nil
}

func (d *fakeDAO) IsDocNotExists(err error) bool {
	return errors.Is(err, errFakeNotExists)
}

// scanInto mirrors the deserialization of the real redis client.
func scanInto(s string, val interface{}) error {
	if u, ok := val.(encoding.BinaryUnmarshaler); ok {
		return u.UnmarshalBinary([]byte(s))
	}

	if p, ok := val.(*string); ok {
		*p = s
		return nil
	}

	return json.Unmarshal([]byte(s), val)
}

func newImpl(d *fakeDAO) *claConfirmTokenImpl {
	return NewCLAConfirmTokenImpl(d)
}

func TestAddGeneratesHexTokenWithTTL(t *testing.T) {
	d := newFakeDAO()
	impl := newImpl(d)

	ttl := 7 * 24 * time.Hour
	token, err := impl.Add("link1", "alice@example.com", "12", ttl)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if len(token) != 64 {
		t.Fatalf("token length = %d, want 64", len(token))
	}
	if !tokenPattern.MatchString(token) {
		t.Fatalf("token is not 64-hex: %s", token)
	}

	if got := d.expiry[tokenKey(token)]; got != ttl {
		t.Fatalf("token key ttl = %v, want %v", got, ttl)
	}
	if got := d.expiry[indexKey("link1", "alice@example.com")]; got != ttl {
		t.Fatalf("index key ttl = %v, want %v", got, ttl)
	}
}

func TestAddThenConsumeReturnsPayload(t *testing.T) {
	d := newFakeDAO()
	impl := newImpl(d)

	before := time.Now().Unix()
	token, err := impl.Add("link1", "alice@example.com", "12", time.Hour)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	payload, err := impl.Consume(token)
	if err != nil {
		t.Fatalf("Consume failed: %v", err)
	}

	if payload.LinkId != "link1" || payload.Email != "alice@example.com" || payload.NewCLAId != "12" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.IssuedAt.Unix() < before {
		t.Fatalf("issued_at = %v, want >= %v", payload.IssuedAt.Unix(), before)
	}
}

func TestAddInvalidatesPreviousToken(t *testing.T) {
	d := newFakeDAO()
	impl := newImpl(d)

	t1, err := impl.Add("link1", "alice@example.com", "12", time.Hour)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	t2, err := impl.Add("link1", "alice@example.com", "13", time.Hour)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if t1 == t2 {
		t.Fatal("two Add calls should generate different tokens")
	}

	if _, err := impl.Consume(t1); !commonRepo.IsErrorResourceNotFound(err) {
		t.Fatalf("consuming the old token should fail with resource-not-found, got: %v", err)
	}

	if _, err := impl.Consume(t2); err != nil {
		t.Fatalf("consuming the new token should succeed, got: %v", err)
	}
}

func TestConsumeIsOneTimeOnly(t *testing.T) {
	d := newFakeDAO()
	impl := newImpl(d)

	token, err := impl.Add("link1", "alice@example.com", "12", time.Hour)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if _, err := impl.Consume(token); err != nil {
		t.Fatalf("first Consume failed: %v", err)
	}

	if _, err := impl.Consume(token); !commonRepo.IsErrorResourceNotFound(err) {
		t.Fatalf("second Consume should fail with resource-not-found, got: %v", err)
	}
}

func TestConcurrentConsumeOnlyOneSucceeds(t *testing.T) {
	d := newFakeDAO()
	impl := newImpl(d)

	token, err := impl.Add("link1", "alice@example.com", "12", time.Hour)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	var wg sync.WaitGroup
	var success int64
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := impl.Consume(token); err == nil {
				atomic.AddInt64(&success, 1)
			}
		}()
	}
	wg.Wait()

	if success != 1 {
		t.Fatalf("concurrent Consume success count = %d, want 1", success)
	}
}

func TestConsumeRejectsBadFormatWithoutTouchingRedis(t *testing.T) {
	d := newFakeDAO()
	impl := newImpl(d)

	badTokens := []string{
		"",
		"short",
		strings64('z'),  // wrong charset
		tokenOfLen(63), // too short
		tokenOfLen(65), // too long
		"ABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD", // uppercase
	}
	for _, token := range badTokens {
		_, err := impl.Consume(token)
		if !domain.IsErrorOf(err, domain.ErrorCodeCLAConfirmTokenInvalid) {
			t.Fatalf("Consume(%q) should return cla_confirm_token_invalid, got: %v", token, err)
		}
	}

	if d.gadCalls != 0 {
		t.Fatalf("format check should reject before touching redis, GetAndDelete called %d times", d.gadCalls)
	}
}

func strings64(r byte) string {
	b := make([]byte, 64)
	for i := range b {
		b[i] = r
	}
	return string(b)
}

func tokenOfLen(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('0' + i%10)
	}
	return string(b)
}

func TestConsumePassesThroughRepoError(t *testing.T) {
	d := newFakeDAO()
	impl := newImpl(d)

	token, err := impl.Add("link1", "alice@example.com", "12", time.Hour)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	expected := errors.New("redis is down")
	d.failOn[tokenKey(token)] = expected

	if _, err := impl.Consume(token); !errors.Is(err, expected) {
		t.Fatalf("Consume should pass through the repo error, got: %v", err)
	}
}

func TestConsumeUnknownToken(t *testing.T) {
	impl := newImpl(newFakeDAO())

	_, err := impl.Consume(tokenOfLen(64))
	if !commonRepo.IsErrorResourceNotFound(err) {
		t.Fatalf("unknown token should fail with resource-not-found, got: %v", err)
	}
}

func TestAddErrors(t *testing.T) {
	t.Run("index get error", func(t *testing.T) {
		d := newFakeDAO()
		d.failOn[indexKey("link1", "alice@example.com")] = errors.New("boom")
		if _, err := newImpl(d).Add("link1", "alice@example.com", "12", time.Hour); err == nil {
			t.Fatal("Add should fail when reading the index key fails")
		}
	})

	t.Run("del old token error", func(t *testing.T) {
		d := newFakeDAO()
		impl := newImpl(d)
		if _, err := impl.Add("link1", "alice@example.com", "12", time.Hour); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
		d.delAll = errors.New("boom")
		if _, err := impl.Add("link1", "alice@example.com", "13", time.Hour); err == nil {
			t.Fatal("Add should fail when deleting the old token fails")
		}
	})

	t.Run("set token key error", func(t *testing.T) {
		d := newFakeDAO()
		impl := newImpl(d)
		if _, err := impl.Add("link1", "alice@example.com", "12", time.Hour); err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// second Add: index read succeeds and returns a token; make the write
		// of the NEW token key fail by failing any set whose key is not the
		// index key after deleting the old one.
		d.failSetExcept = indexKey("link1", "alice@example.com")
		d.failSetErr = errors.New("write boom")
		if _, err := impl.Add("link1", "alice@example.com", "13", time.Hour); err == nil {
			t.Fatal("Add should fail when writing the token key fails")
		}
	})

	t.Run("set index key error", func(t *testing.T) {
		d := newFakeDAO()
		impl := newImpl(d)
		if _, err := impl.Add("link1", "alice@example.com", "12", time.Hour); err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// second Add: delete old token ok, set new token ok, set index fails
		d.failOn[indexKey("link1", "alice@example.com")] = errors.New("write boom")
		if _, err := impl.Add("link1", "alice@example.com", "13", time.Hour); err == nil {
			t.Fatal("Add should fail when writing the index key fails")
		}
	})
}

var _ = repository.CLAConfirmToken(&claConfirmTokenImpl{})
