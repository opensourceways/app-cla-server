package repositoryimpl

import (
	"testing"
)

func TestGenDoc(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	doc := testStruct{Name: "test", Value: 42}
	result, err := genDoc(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("expected 'test', got '%v'", result["name"])
	}
	if v, ok := result["value"].(float64); !ok || int(v) != 42 {
		t.Errorf("expected 42, got %v", result["value"])
	}
}

func TestLinkIdFilter(t *testing.T) {
	filter := linkIdFilter("link-1")
	if filter[fieldLinkId] != "link-1" {
		t.Errorf("expected 'link-1', got '%v'", filter[fieldLinkId])
	}
}

func TestChildField(t *testing.T) {
	result := childField("a", "b", "c")
	if result != "a.b.c" {
		t.Errorf("expected 'a.b.c', got '%s'", result)
	}

	result = childField("single")
	if result != "single" {
		t.Errorf("expected 'single', got '%s'", result)
	}
}

func TestMongodbConstants(t *testing.T) {
	if mongodbCmdOr != "$or" || mongodbCmdIn != "$in" {
		t.Error("constant mismatch")
	}
	if mongodbCmdLt != "$lt" || mongodbCmdElemMatch != "$elemMatch" {
		t.Error("constant mismatch")
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Collections: Collections{
			Org:               "org",
			CLA:               "cla",
			Link:              "link",
			User:              "user",
			PrivacyConsent:    "privacy",
			CorpSigning:       "corp_signing",
			EmailCredential:   "email_credential",
			VerificationCode:  "verification_code",
			IndividualSigning: "individual_signing",
		},
	}
	if cfg.Collections.Org != "org" || cfg.Collections.IndividualSigning != "individual_signing" {
		t.Errorf("config mismatch")
	}
}
