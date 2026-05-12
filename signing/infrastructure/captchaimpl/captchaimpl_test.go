package captchaimpl

import (
	"testing"
)

func TestCaptchaImplStruct(t *testing.T) {
	c := &captchaImpl{
		isTestEnv:  true,
		testAnswer: "123456",
	}
	if !c.isTestEnv {
		t.Error("expected isTestEnv=true")
	}
	if c.testAnswer != "123456" {
		t.Errorf("expected '123456', got '%s'", c.testAnswer)
	}
}

func TestVerifyEmptyInput(t *testing.T) {
	c := &captchaImpl{}
	err := c.Verify("", "")
	if err == nil {
		t.Error("expected error for empty input")
	}
	err = c.Verify("id", "")
	if err == nil {
		t.Error("expected error for empty answer")
	}
	err = c.Verify("", "answer")
	if err == nil {
		t.Error("expected error for empty id")
	}
}
