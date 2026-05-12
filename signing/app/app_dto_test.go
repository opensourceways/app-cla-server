package app

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestCLAInfoToCLA(t *testing.T) {
	url, _ := dp.NewURL("https://example.com/cla.pdf")
	cmd := &CLAInfo{
		URL:      url,
		Text:     []byte("text"),
		Type:     dp.CLATypeCorp,
		Fields:   []domain.Field{},
		Language: dp.CreateLanguage("en"),
	}
	cla := cmd.toCLA()
	if cla.URL.URL() != "https://example.com/cla.pdf" {
		t.Errorf("URL mismatch")
	}
	if cla.Type.CLAType() != "corporation" {
		t.Errorf("type mismatch")
	}
}

func TestCmdToUpdateCLANewCLA(t *testing.T) {
	url, _ := dp.NewURL("https://example.com/new.pdf")
	cmd := &CmdToUpdateCLA{
		URL:      url,
		Text:     []byte("new text"),
		Type:     dp.CLATypeIndividual,
		Language: dp.CreateLanguage("en"),
	}
	cla := cmd.newCLA()
	if cla.URL.URL() != "https://example.com/new.pdf" {
		t.Errorf("URL mismatch")
	}
	if cla.Type.CLAType() != "individual" {
		t.Errorf("type mismatch")
	}
}

func TestCmdToAddLinkToLink(t *testing.T) {
	url, _ := dp.NewURL("https://example.com/c.pdf")
	email, _ := dp.NewEmailAddr("org@test.com")
	cmd := &CmdToAddLink{
		Org: domain.OrgInfo{
			Alias:      "TestOrg",
			Logo:       "logo.png",
			ProjectURL: "https://example.com",
		},
		Email:     email,
		Submitter: "admin",
		CLAs: []CLAInfo{
			{
				URL:      url,
				Text:     []byte("CLA content"),
				Type:     dp.CLATypeCorp,
				Language: dp.CreateLanguage("en"),
			},
		},
	}
	link := cmd.toLink()
	if link.Org.Alias != "TestOrg" {
		t.Errorf("expected 'TestOrg', got '%s'", link.Org.Alias)
	}
	if link.Submitter != "admin" {
		t.Errorf("expected 'admin', got '%s'", link.Submitter)
	}
	if len(link.CLAs) != 1 {
		t.Fatalf("expected 1 CLA, got %d", len(link.CLAs))
	}
	if link.CLAs[0].Id != "0" {
		t.Errorf("expected Id '0', got '%s'", link.CLAs[0].Id)
	}
	if link.CLANum != 1 {
		t.Errorf("expected CLANum=1, got %d", link.CLANum)
	}
}

func TestCmdToAddLinkWithMultipleCLAs(t *testing.T) {
	url1, _ := dp.NewURL("https://example.com/a.pdf")
	url2, _ := dp.NewURL("https://example.com/b.pdf")
	email, _ := dp.NewEmailAddr("org@test.com")
	cmd := &CmdToAddLink{
		Email:     email,
		Submitter: "admin",
		Org:       domain.OrgInfo{Alias: "X"},
		CLAs: []CLAInfo{
			{URL: url1, Text: []byte("a"), Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
			{URL: url2, Text: []byte("b"), Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")},
		},
	}
	link := cmd.toLink()
	if len(link.CLAs) != 2 {
		t.Fatalf("expected 2 CLAs, got %d", len(link.CLAs))
	}
	if link.CLAs[0].Id != "0" || link.CLAs[1].Id != "1" {
		t.Errorf("expected Ids 0 and 1, got %s and %s", link.CLAs[0].Id, link.CLAs[1].Id)
	}
	if link.CLANum != 2 {
		t.Errorf("expected CLANum=2, got %d", link.CLANum)
	}
}
