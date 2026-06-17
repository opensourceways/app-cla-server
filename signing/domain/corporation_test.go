package domain

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestNewCorporation(t *testing.T) {
	email := dp.CreateEmailAddr("info@testcorp.com")
	name := dp.CreateCorpName("TestCorp")

	corp := NewCorporation(name, email)

	if corp.Name.CorpName() != "TestCorp" {
		t.Errorf("Name: got %s, want TestCorp", corp.Name.CorpName())
	}
	if corp.PrimaryEmailDomain != "testcorp.com" {
		t.Errorf("PrimaryEmailDomain: got %s, want testcorp.com", corp.PrimaryEmailDomain)
	}
	if len(corp.AllEmailDomains) != 1 || corp.AllEmailDomains[0] != "testcorp.com" {
		t.Errorf("AllEmailDomains: got %v", corp.AllEmailDomains)
	}
}

func TestCorporationIsMyEmail(t *testing.T) {
	corp := Corporation{
		AllEmailDomains:    []string{"corp.com", "other.com"},
		PrimaryEmailDomain: "corp.com",
	}

	if !corp.isMyEmail(dp.CreateEmailAddr("user@corp.com")) {
		t.Error("isMyEmail should match corp.com")
	}
	if !corp.isMyEmail(dp.CreateEmailAddr("user@other.com")) {
		t.Error("isMyEmail should match other.com")
	}
	if corp.isMyEmail(dp.CreateEmailAddr("user@unknown.com")) {
		t.Error("isMyEmail should not match unknown.com")
	}
}

func TestCorporationAddEmailDomain(t *testing.T) {
	corp := Corporation{AllEmailDomains: []string{"corp.com"}}

	if err := corp.addEmailDomain("new.com"); err != nil {
		t.Fatalf("addEmailDomain new: unexpected error %v", err)
	}
	if len(corp.AllEmailDomains) != 2 {
		t.Errorf("AllEmailDomains length: got %d, want 2", len(corp.AllEmailDomains))
	}

	if err := corp.addEmailDomain("corp.com"); err == nil {
		t.Error("addEmailDomain duplicate should fail")
	}
}

func TestCorpSigningAddEmailDomain(t *testing.T) {
	cs := &CorpSigning{Corp: Corporation{AllEmailDomains: []string{"corp.com"}}}

	if err := cs.AddEmailDomain(dp.CreateEmailAddr("user@new.com")); err != nil {
		t.Fatalf("AddEmailDomain new: unexpected error %v", err)
	}
	if err := cs.AddEmailDomain(dp.CreateEmailAddr("user@corp.com")); err == nil {
		t.Error("AddEmailDomain duplicate should fail")
	}
}
