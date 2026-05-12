package util

import (
	"testing"
)

func TestHasXSS(t *testing.T) {
	if !HasXSS("<script>") {
		t.Error("expected true for XSS content '<script>'")
	}
	if !HasXSS("a&b") {
		t.Error("expected true for XSS content 'a&b'")
	}
	if !HasXSS(`"quote"`) {
		t.Error("expected true for XSS content with double quote")
	}
	if !HasXSS("a'b") {
		t.Error("expected true for XSS content with single quote")
	}
	if !HasXSS("a/b") {
		t.Error("expected true for XSS content with slash")
	}
	if HasXSS("hello world") {
		t.Error("expected false for normal text")
	}
	if HasXSS("a-b_c.123") {
		t.Error("expected false for safe characters")
	}
}

func TestCheckEmail(t *testing.T) {
	if err := CheckEmail("test@example.com"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
	if err := CheckEmail("user+tag@example.co.uk"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
	if err := CheckEmail("a@b.cd"); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}

	if err := CheckEmail(""); err == nil {
		t.Error("expected invalid for empty email")
	}
	if err := CheckEmail("not-an-email"); err == nil {
		t.Error("expected invalid for 'not-an-email'")
	}
	if err := CheckEmail("@missing-user.com"); err == nil {
		t.Error("expected invalid for '@missing-user.com'")
	}
}

func TestStrLen(t *testing.T) {
	if n := StrLen("hello"); n != 5 {
		t.Errorf("expected 5, got %d", n)
	}
	if n := StrLen("你好世界"); n != 4 {
		t.Errorf("expected 4, got %d", n)
	}
	if n := StrLen(""); n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestEmailSuffix(t *testing.T) {
	if s := EmailSuffix("user@example.com"); s != "example.com" {
		t.Errorf("expected 'example.com', got '%s'", s)
	}
	if s := EmailSuffix("a@b.c"); s != "b.c" {
		t.Errorf("expected 'b.c', got '%s'", s)
	}
	if s := EmailSuffix("no-at-sign"); s != "no-at-sign" {
		t.Errorf("expected 'no-at-sign', got '%s'", s)
	}
	if s := EmailSuffix(""); s != "" {
		t.Errorf("expected '', got '%s'", s)
	}
}

func TestGenFilePath(t *testing.T) {
	p := GenFilePath("/tmp", "file.txt")
	if p != "/tmp/file.txt" {
		t.Errorf("expected '/tmp/file.txt', got '%s'", p)
	}
}
