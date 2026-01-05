package vcservice

import (
	"testing"
	"time"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/randomcode"
)

// mockVerificationCodeRepo 模拟验证码存储
type mockVerificationCodeRepo struct {
	codes map[string]domain.VerificationCode // 注意：存储值而不是指针
}

func newMockVerificationCodeRepo() *mockVerificationCodeRepo {
	return &mockVerificationCodeRepo{
		codes: make(map[string]domain.VerificationCode),
	}
}

func (m *mockVerificationCodeRepo) Add(vc *domain.VerificationCode) error {
	key := vc.Purpose.Purpose()
	m.codes[key] = *vc // 存储值的副本
	return nil
}

// 修复：返回值而不是指针
func (m *mockVerificationCodeRepo) Find(key *domain.VerificationCodeKey) (domain.VerificationCode, error) {
	k := key.Purpose.Purpose()
	if vc, exists := m.codes[k]; exists {
		return vc, nil // 返回值而不是指针
	}
	return domain.VerificationCode{}, domain.NewNotFoundDomainError("not_found")
}

// mockLimiter 模拟限流器
type mockLimiter struct{}

func (m *mockLimiter) IsAllowed(purpose string) (bool, error) {
	return true, nil
}

func (m *mockLimiter) Add(purpose string, interval time.Duration) error {
	return nil
}

func TestVCService_WithTestRandomCode(t *testing.T) {
	// 创建测试环境的验证码服务
	vcService := NewVCService(
		newMockVerificationCodeRepo(),
		&mockLimiter{},
		randomcode.NewTestRandomCode("000000"),
	)

	// 创建测试用的 purpose
	purpose, err := dp.NewPurpose("test, individual, test@example.com")
	if err != nil {
		t.Fatalf("Failed to create purpose: %v", err)
	}

	// 测试验证码生成
	code, err := vcService.New(purpose)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if code != "000000" {
		t.Errorf("Expected '000000', got '%s'", code)
	}

	// 测试验证码验证
	key := domain.NewVerificationCodeKey("000000", purpose)
	err = vcService.Verify(&key)
	if err != nil {
		t.Errorf("Expected valid verification, got error: %v", err)
	}

	// 测试错误验证码
	wrongKey := domain.NewVerificationCodeKey("000100", purpose)
	err = vcService.Verify(&wrongKey)
	if err == nil {
		t.Error("Expected error for wrong verification code")
	}
}

func TestVCService_NewIfItCan_WithTestRandomCode(t *testing.T) {
	vcService := NewVCService(
		newMockVerificationCodeRepo(),
		&mockLimiter{},
		randomcode.NewTestRandomCode("000000"),
	)

	purpose, err := dp.NewPurpose("test, corp, test@example.com")
	if err != nil {
		t.Fatalf("Failed to create purpose: %v", err)
	}

	// 测试带限流的验证码生成
	code, err := vcService.NewIfItCan(purpose, time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if code != "000000" {
		t.Errorf("Expected '000000', got '%s'", code)
	}
}
