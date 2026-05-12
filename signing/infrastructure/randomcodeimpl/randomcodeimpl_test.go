package randomcodeimpl

import (
	"testing"
)

func TestIsValid(t *testing.T) {
	impl := NewRandomCodeImpl()

	if !impl.IsValid("123456") {
		t.Error("expected valid for '123456'")
	}
	if !impl.IsValid("000000") {
		t.Error("expected valid for '000000'")
	}
	if !impl.IsValid("999999") {
		t.Error("expected valid for '999999'")
	}

	if impl.IsValid("12345") {
		t.Error("expected invalid for length 5")
	}
	if impl.IsValid("1234567") {
		t.Error("expected invalid for length 7")
	}
	if impl.IsValid("abcdef") {
		t.Error("expected invalid for letters")
	}
	if impl.IsValid("") {
		t.Error("expected invalid for empty")
	}
	if impl.IsValid("1234 6") {
		t.Error("expected invalid for space")
	}
}

func TestNew(t *testing.T) {
	impl := NewRandomCodeImpl()
	code, err := impl.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !impl.IsValid(code) {
		t.Errorf("generated code '%s' should be valid", code)
	}
}
