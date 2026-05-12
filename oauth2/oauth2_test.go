package oauth2

import (
	"testing"
)

func TestBuildOauth2Config(t *testing.T) {
	cfg := Oauth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		AuthURL:      "https://auth.example.com",
		TokenURL:     "https://token.example.com",
		RedirectURL:  "https://redirect.example.com",
		Scope:        []string{"user", "email"},
	}

	result := buildOauth2Config(cfg)

	if result.ClientID != "test-client-id" {
		t.Errorf("expected 'test-client-id', got '%s'", result.ClientID)
	}
	if result.ClientSecret != "test-secret" {
		t.Errorf("expected 'test-secret', got '%s'", result.ClientSecret)
	}
	if result.Endpoint.AuthURL != "https://auth.example.com" {
		t.Errorf("expected 'https://auth.example.com', got '%s'", result.Endpoint.AuthURL)
	}
	if result.Endpoint.TokenURL != "https://token.example.com" {
		t.Errorf("expected 'https://token.example.com', got '%s'", result.Endpoint.TokenURL)
	}
	if result.RedirectURL != "https://redirect.example.com" {
		t.Errorf("expected 'https://redirect.example.com', got '%s'", result.RedirectURL)
	}
	if len(result.Scopes) != 2 || result.Scopes[0] != "user" || result.Scopes[1] != "email" {
		t.Errorf("expected scopes [user email], got %v", result.Scopes)
	}
}

func TestNewOauth2Client(t *testing.T) {
	cfg := Oauth2Config{
		ClientID:     "id",
		ClientSecret: "secret",
		AuthURL:      "https://auth.example.com",
		TokenURL:     "https://token.example.com",
		RedirectURL:  "https://redirect.example.com",
		Scope:        []string{"user"},
	}
	client := NewOauth2Client(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
