package app

import (
	"testing"
	"time"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type mockVCService struct {
	newCode        string
	newErr         error
	newIfItCanCode string
	newIfItCanErr  error
	verifyErr      error
}

func (m *mockVCService) New(p dp.Purpose) (string, error)            { return m.newCode, m.newErr }
func (m *mockVCService) NewIfItCan(p dp.Purpose, interval time.Duration) (string, error) {
	return m.newIfItCanCode, m.newIfItCanErr
}
func (m *mockVCService) Verify(key *domain.VerificationCodeKey) error { return m.verifyErr }

type mockVCPurpose struct {
	p   dp.Purpose
	err error
}

func (m *mockVCPurpose) purpose() (dp.Purpose, error) { return m.p, m.err }

func TestVerificationCodeServiceNewCode(t *testing.T) {
	vc := &verificationCodeService{
		vc: &mockVCService{newCode: "123456"},
	}
	p, _ := dp.NewPurpose("test-purpose")
	code, err := vc.newCode(&mockVCPurpose{p: p})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "123456" {
		t.Errorf("expected '123456', got '%s'", code)
	}
}

func TestVerificationCodeServiceNewCodeIfItCan(t *testing.T) {
	vc := &verificationCodeService{
		vc: &mockVCService{newIfItCanCode: "654321"},
	}
	p, _ := dp.NewPurpose("test-purpose")
	code, err := vc.newCodeIfItCan(&mockVCPurpose{p: p}, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "654321" {
		t.Errorf("expected '654321', got '%s'", code)
	}
}

func TestVerificationCodeServiceValidate(t *testing.T) {
	vc := &verificationCodeService{
		vc: &mockVCService{},
	}
	p, _ := dp.NewPurpose("validate-purpose")
	err := vc.validate(&mockVCPurpose{p: p}, "code123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
