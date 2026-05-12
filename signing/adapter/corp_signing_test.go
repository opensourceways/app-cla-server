package adapter

import (
	"testing"
)

func TestIsValidaCorpEmailDomain(t *testing.T) {
	adapter := &corpSigningAdatper{
		invalidCorpEmailDomain: []string{"gmail.com", "yahoo.com"},
	}

	if !adapter.isValidaCorpEmailDomain("corp.com") {
		t.Error("expected valid for non-blocklisted domain")
	}
	if adapter.isValidaCorpEmailDomain("gmail.com") {
		t.Error("expected invalid for gmail.com")
	}
	if adapter.isValidaCorpEmailDomain("GMAIL.com") {
		t.Error("expected invalid for GMAIL.com (case insensitive)")
	}
}

func TestIsValidaCorpEmailDomainEmptyBlacklist(t *testing.T) {
	adapter := &corpSigningAdatper{
		invalidCorpEmailDomain: []string{},
	}
	if !adapter.isValidaCorpEmailDomain("gmail.com") {
		t.Error("expected valid when blacklist is empty")
	}
}

func TestNewCorpSigningAdapter(t *testing.T) {
	adapter := NewCorpSigningAdapter(nil, []string{"GMAIL.COM", "YAHOO.COM"})
	if len(adapter.invalidCorpEmailDomain) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(adapter.invalidCorpEmailDomain))
	}
	if adapter.invalidCorpEmailDomain[0] != "gmail.com" {
		t.Errorf("expected lowercase 'gmail.com', got '%s'", adapter.invalidCorpEmailDomain[0])
	}
	if adapter.invalidCorpEmailDomain[1] != "yahoo.com" {
		t.Errorf("expected lowercase 'yahoo.com', got '%s'", adapter.invalidCorpEmailDomain[1])
	}
}

func TestNewCorpSigningAdapterEmptyDomains(t *testing.T) {
	adapter := NewCorpSigningAdapter(nil, nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
	if len(adapter.invalidCorpEmailDomain) != 0 {
		t.Errorf("expected 0 domains, got %d", len(adapter.invalidCorpEmailDomain))
	}
}
