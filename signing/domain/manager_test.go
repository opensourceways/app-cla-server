package domain

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestManagerIsEmpty(t *testing.T) {
	m1 := Manager{Id: ""}
	if !m1.IsEmpty() {
		t.Error("IsEmpty should return true for empty Id")
	}

	m2 := Manager{Id: "some_id"}
	if m2.IsEmpty() {
		t.Error("IsEmpty should return false for non-empty Id")
	}
}

func TestManagerIsMe(t *testing.T) {
	email := dp.CreateEmailAddr("admin@test.com")
	m := Manager{Id: "1", Representative: Representative{EmailAddr: email}}

	if !m.isMe(email) {
		t.Error("isMe should return true for same email")
	}

	other := dp.CreateEmailAddr("other@test.com")
	if m.isMe(other) {
		t.Error("isMe should return false for different email")
	}

	mEmpty := Manager{Id: "", Representative: Representative{EmailAddr: email}}
	if mEmpty.isMe(email) {
		t.Error("isMe should return false for empty Id")
	}
}

func TestManagerIsSame(t *testing.T) {
	m1 := &Manager{Id: "1", Representative: Representative{EmailAddr: dp.CreateEmailAddr("a@test.com")}}
	m2 := &Manager{Id: "2", Representative: Representative{EmailAddr: dp.CreateEmailAddr("a@test.com")}}
	m3 := &Manager{Id: "1", Representative: Representative{EmailAddr: dp.CreateEmailAddr("b@test.com")}}

	if !m1.IsSame(m2) {
		t.Error("IsSame should return true for same email")
	}
	if !m1.IsSame(m3) {
		t.Error("IsSame should return true for same Id")
	}

	m4 := &Manager{Id: "4", Representative: Representative{EmailAddr: dp.CreateEmailAddr("b@test.com")}}
	if m1.IsSame(m4) {
		t.Error("IsSame should return false for different Id and email")
	}
}

func TestManagerHasEmail(t *testing.T) {
	email := dp.CreateEmailAddr("admin@test.com")
	m := Manager{Representative: Representative{EmailAddr: email}}

	if !m.hasEmail(email) {
		t.Error("hasEmail should return true for same email")
	}

	other := dp.CreateEmailAddr("other@test.com")
	if m.hasEmail(other) {
		t.Error("hasEmail should return false for different email")
	}
}

func TestManagerAccount(t *testing.T) {
	m := Manager{Id: "user1", Representative: Representative{EmailAddr: dp.CreateEmailAddr("u@domain.com")}}
	acc, err := m.Account()
	if err != nil {
		t.Fatalf("Account: unexpected error %v", err)
	}
	if acc.Account() != "user1_domain.com" {
		t.Errorf("Account: got %s, want user1_domain.com", acc.Account())
	}

	m2 := Manager{Id: ""}
	_, err2 := m2.Account()
	if err2 == nil {
		t.Error("Account for empty manager should fail")
	}
}

func TestNewRepresentative(t *testing.T) {
	r, err := NewRepresentative("John", "john@example.com")
	if err != nil {
		t.Fatalf("NewRepresentative valid: unexpected error %v", err)
	}
	if r.Name.Name() != "John" {
		t.Errorf("Name: got %s, want John", r.Name.Name())
	}
	if r.EmailAddr.EmailAddr() != "john@example.com" {
		t.Errorf("EmailAddr: got %s, want john@example.com", r.EmailAddr.EmailAddr())
	}

	_, err = NewRepresentative("", "john@example.com")
	if err == nil {
		t.Error("NewRepresentative with empty name should fail")
	}

	_, err = NewRepresentative("John", "invalid-email")
	if err == nil {
		t.Error("NewRepresentative with invalid email should fail")
	}
}
