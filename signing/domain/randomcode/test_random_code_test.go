package randomcode

import (
	"testing"
)

func TestTestRandomCode_New(t *testing.T) {
	// 创建测试环境的验证码生成器
	testCode := NewTestRandomCode("000000")

	// 多次调用验证码生成，应该始终返回 "123456"
	for i := 0; i < 5; i++ {
		code, err := testCode.New()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if code != "000000" {
			t.Errorf("Expected '123456', got '%s'", code)
		}
	}
}

func TestTestRandomCode_IsValid(t *testing.T) {
	testCode := NewTestRandomCode("000000")

	// 测试有效验证码
	if !testCode.IsValid("000000") {
		t.Error("Expected '000000' to be valid")
	}

	// 测试无效验证码
	invalidCodes := []string{"12345", "1234567", "000100", "abcdef", ""}
	for _, code := range invalidCodes {
		if testCode.IsValid(code) {
			t.Errorf("Expected '%s' to be invalid", code)
		}
	}
}

func TestTestRandomCode_Interface(t *testing.T) {
	// 确保 TestRandomCode 实现了 RandomCode 接口
	var _ RandomCode = NewTestRandomCode("000000")
}
