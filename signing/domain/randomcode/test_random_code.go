package randomcode

// TestRandomCode 固定返回配置的验证码实现，用于测试环境
type TestRandomCode struct {
	code string
}

func NewTestRandomCode(testCode string) RandomCode {
	if testCode == "" {
		testCode = "123456" // 默认值
	}
	return &TestRandomCode{
		code: testCode,
	}
}

func (t *TestRandomCode) New() (string, error) {
	return t.code, nil
}

func (t *TestRandomCode) IsValid(code string) bool {
	return code == t.code
}
