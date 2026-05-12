package adapter

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/app"
)

func TestToCorporationManagerCreateOption(t *testing.T) {
	dto := &app.ManagerDTO{
		Account:   "user001",
		Role:      "admin",
		Name:      "John",
		EmailAddr: "john@test.com",
		Password:  []byte("secret"),
	}
	result := toCorporationManagerCreateOption(dto)
	if result.ID != "user001" {
		t.Errorf("expected 'user001', got '%s'", result.ID)
	}
	if result.Role != "admin" {
		t.Errorf("expected 'admin', got '%s'", result.Role)
	}
	if result.Name != "John" {
		t.Errorf("expected 'John', got '%s'", result.Name)
	}
	if result.Email != "john@test.com" {
		t.Errorf("expected 'john@test.com', got '%s'", result.Email)
	}
	if string(result.Password) != "secret" {
		t.Errorf("expected 'secret', got '%s'", string(result.Password))
	}
}

func TestToCorporationManagerListResult(t *testing.T) {
	dto := &app.EmployeeManagerDTO{
		ID:    "manager-1",
		Name:  "Alice",
		Email: "alice@test.com",
	}
	result := toCorporationManagerListResult(dto)
	if result.ID != "manager-1" {
		t.Errorf("expected 'manager-1', got '%s'", result.ID)
	}
	if result.Name != "Alice" {
		t.Errorf("expected 'Alice', got '%s'", result.Name)
	}
	if result.Email != "alice@test.com" {
		t.Errorf("expected 'alice@test.com', got '%s'", result.Email)
	}
}

func TestNewEmployeeManagerAdapter(t *testing.T) {
	adapter := NewEmployeeManagerAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewEmployeeSigningAdapter(t *testing.T) {
	adapter := NewEmployeeSigningAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewCorpEmailDomainAdapter(t *testing.T) {
	adapter := NewCorpEmailDomainAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewCorpAdminAdapter(t *testing.T) {
	adapter := NewCorpAdminAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewCorpPDFAdapter(t *testing.T) {
	adapter := NewCorpPDFAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewIndividualSigningAdapter(t *testing.T) {
	adapter := NewIndividualSigningAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}
