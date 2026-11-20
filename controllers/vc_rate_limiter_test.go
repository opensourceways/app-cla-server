package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	beecontext "github.com/beego/beego/v2/server/web/context"
	"github.com/opensourceways/obs-sdk/go/sdkctx"
)

type mockVCRateLimiter struct {
	allowResult bool
	allowErr    error
	called      bool
	calledIP    string
}

func (m *mockVCRateLimiter) Allow(ip string) (bool, error) {
	m.called = true
	m.calledIP = ip
	return m.allowResult, m.allowErr
}

func newBaseControllerWithIP(ip string) *baseController {
	ctx := beecontext.NewContext()

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/v1/link/abc/code", nil)
	inner := sdkctx.WithRequestID(context.Background(), "req-test")
	r = r.WithContext(inner)

	if ip != "" {
		r.Header.Set("x-forwarded-for", ip)
	}

	ctx.Request = r
	ctx.ResponseWriter = &beecontext.Response{
		ResponseWriter: recorder,
	}
	ctx.Input.Reset(ctx)
	ctx.Output.Context = ctx

	ctl := &baseController{}
	ctl.Ctx = ctx
	ctl.Data = map[interface{}]interface{}{}

	return ctl
}

func TestCheckVCRateLimitNilLimiter(t *testing.T) {
	orig := vcLimiter
	vcLimiter = nil
	defer func() { vcLimiter = orig }()

	ctl := newBaseControllerWithIP("1.2.3.4")
	fr := ctl.checkVCRateLimit()
	if fr != nil {
		t.Errorf("expected nil when vcLimiter is nil, got %v", fr)
	}
}

func TestCheckVCRateLimitAllowed(t *testing.T) {
	orig := vcLimiter
	defer func() { vcLimiter = orig }()

	mock := &mockVCRateLimiter{allowResult: true}
	vcLimiter = mock

	ctl := newBaseControllerWithIP("10.0.0.1")
	fr := ctl.checkVCRateLimit()
	if fr != nil {
		t.Errorf("expected nil when allowed, got %v", fr)
	}
	if !mock.called {
		t.Error("expected Allow to be called")
	}
	if mock.calledIP != "10.0.0.1" {
		t.Errorf("expected IP 10.0.0.1, got %s", mock.calledIP)
	}
}

func TestCheckVCRateLimitDenied(t *testing.T) {
	orig := vcLimiter
	defer func() { vcLimiter = orig }()

	mock := &mockVCRateLimiter{allowResult: false}
	vcLimiter = mock

	ctl := newBaseControllerWithIP("10.0.0.2")
	fr := ctl.checkVCRateLimit()
	if fr == nil {
		t.Fatal("expected failed result when denied")
	}
	if fr.errCode != errTooManyRequest {
		t.Errorf("errCode: got %q, want %q", fr.errCode, errTooManyRequest)
	}
	if fr.statusCode != 400 {
		t.Errorf("statusCode: got %d, want 400", fr.statusCode)
	}
}

func TestCheckVCRateLimitLimiterError(t *testing.T) {
	orig := vcLimiter
	defer func() { vcLimiter = orig }()

	mock := &mockVCRateLimiter{allowErr: errors.New("redis down")}
	vcLimiter = mock

	ctl := newBaseControllerWithIP("10.0.0.3")
	fr := ctl.checkVCRateLimit()
	if fr == nil {
		t.Fatal("expected failed result when limiter errors")
	}
	if fr.errCode != errTooManyRequest {
		t.Errorf("errCode: got %q, want %q", fr.errCode, errTooManyRequest)
	}
}

func TestCheckVCRateLimitNoIPHeader(t *testing.T) {
	orig := vcLimiter
	defer func() { vcLimiter = orig }()

	mock := &mockVCRateLimiter{allowResult: true}
	vcLimiter = mock

	ctl := newBaseControllerWithIP("")
	fr := ctl.checkVCRateLimit()
	if fr == nil {
		t.Fatal("expected failed result when IP cannot be fetched")
	}
	if fr.errCode != errCanNotFetchClientIP {
		t.Errorf("errCode: got %q, want %q", fr.errCode, errCanNotFetchClientIP)
	}
	if mock.called {
		t.Error("Allow should not be called when IP is missing")
	}
}
