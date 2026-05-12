package util

import (
	"testing"
)

func TestDate(t *testing.T) {
	d := Date()
	if len(d) != 10 {
		t.Errorf("expected YYYY-MM-DD format, got '%s'", d)
	}
}

func TestTime(t *testing.T) {
	tm := Time()
	if len(tm) != 19 {
		t.Errorf("expected YYYY-MM-DD HH:MM:SS format, got '%s'", tm)
	}
}

func TestNow(t *testing.T) {
	n := Now()
	if n <= 0 {
		t.Errorf("expected positive timestamp, got %d", n)
	}
}

func TestExpiry(t *testing.T) {
	n := Now()
	e := Expiry(10) // 10 seconds
	if e <= n {
		t.Errorf("expected expiry > now, got expiry=%d now=%d", e, n)
	}
}

func TestCheckContentType(t *testing.T) {
	if !CheckContentType([]byte("<html>"), "text/html") {
		t.Error("expected true for HTML content type")
	}
	if CheckContentType([]byte("<html>"), "image/png") {
		t.Error("expected false for mismatched content type")
	}

	// plain text detection
	data := []byte("hello world")
	if !CheckContentType(data, "text/plain") {
		t.Error("expected true for plain text content")
	}
}
