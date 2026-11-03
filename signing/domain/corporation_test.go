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
	corp := Corporation{
		AllEmailDomains:    []string{"corp.com"},
		PrimaryEmailDomain: "corp.com",
	}

	if err := corp.addEmailDomain("sub.corp.com"); err != nil {
		t.Fatalf("addEmailDomain subdomain: unexpected error %v", err)
	}
	if len(corp.AllEmailDomains) != 2 {
		t.Errorf("AllEmailDomains length: got %d, want 2", len(corp.AllEmailDomains))
	}

	if err := corp.addEmailDomain("corp.com"); err == nil {
		t.Error("addEmailDomain duplicate should fail")
	}

	if err := corp.addEmailDomain("bbb.com"); err == nil {
		t.Error("addEmailDomain unrelated domain should fail")
	}
}

func TestCorpSigningAddEmailDomain(t *testing.T) {
	cs := &CorpSigning{Corp: Corporation{
		AllEmailDomains:    []string{"corp.com"},
		PrimaryEmailDomain: "corp.com",
	}}

	if err := cs.AddEmailDomain(dp.CreateEmailAddr("user@sub.corp.com")); err != nil {
		t.Fatalf("AddEmailDomain subdomain: unexpected error %v", err)
	}
	if err := cs.AddEmailDomain(dp.CreateEmailAddr("user@corp.com")); err == nil {
		t.Error("AddEmailDomain duplicate should fail")
	}
	if err := cs.AddEmailDomain(dp.CreateEmailAddr("user@bbb.com")); err == nil {
		t.Error("AddEmailDomain unrelated domain should fail")
	}
}

func TestCorporationAddEmailDomainSubdomainValidation(t *testing.T) {
	corp := Corporation{
		AllEmailDomains:    []string{"aaa.com"},
		PrimaryEmailDomain: "aaa.com",
	}

	cases := []struct {
		name    string
		domain  string
		wantErr bool
		errCode string
	}{
		{"valid subdomain", "xxx.aaa.com", false, ""},
		{"valid deep subdomain", "a.b.aaa.com", false, ""},
		{"unrelated domain", "bbb.com", true, ErrorCodeCorpEmailDomainNotMatch},
		{"look-alike domain", "baaa.com", true, ErrorCodeCorpEmailDomainNotMatch},
		{"sibling tld", "aaa.org", true, ErrorCodeCorpEmailDomainNotMatch},
		{"case-insensitive subdomain", "SUB.AAA.COM", false, ""},
		{"primary itself is duplicate", "aaa.com", true, ErrorCodeCorpEmailDomainExists},
		{"already added subdomain", "xxx.aaa.com", true, ErrorCodeCorpEmailDomainExists},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := corp.addEmailDomain(tc.domain)
			if tc.wantErr {
				if err == nil {
					t.Errorf("addEmailDomain(%q): expected error, got nil", tc.domain)
					return
				}
				if tc.errCode != "" && !IsErrorOf(err, tc.errCode) {
					t.Errorf("addEmailDomain(%q): error code got %v, want %s", tc.domain, err, tc.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("addEmailDomain(%q): unexpected error %v", tc.domain, err)
				}
			}
		})
	}
}

func TestCorporationAddEmailDomainNoPrimary(t *testing.T) {
	corp := Corporation{AllEmailDomains: []string{"corp.com"}}

	if err := corp.addEmailDomain("sub.corp.com"); err == nil {
		t.Error("addEmailDomain without primary domain should fail")
	}
}
