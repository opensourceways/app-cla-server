package adapter

import (
	"errors"
	"testing"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

// ---- mock ----

type mockCorpSigningService struct {
	app.CorpSigningService // embed nil interface

	created *app.ManagerDTO
	err     error

	calledRep  bool
	repName    string
	repEmail   string
}

func (m *mockCorpSigningService) UpdateRepresentative(userId, linkID, signingID, repName, repEmail string) (*app.ManagerDTO, error) {
	m.calledRep = true
	m.repName = repName
	m.repEmail = repEmail
	return m.created, m.err
}

// ---- tests ----

func TestUpdateRepresentative_Validation_EmptyName(t *testing.T) {
	a := &corpSigningAdatper{s: &mockCorpSigningService{}}
	opt := &models.RepresentativeUpdateOption{RepName: "", RepEmail: "b@corp.com"}

	r, merr := a.UpdateRepresentative("u", "l", "s", opt)
	if merr == nil {
		t.Fatal("expected model error for empty name")
	}
	if r != nil {
		t.Errorf("result should be nil, got %+v", r)
	}
}

func TestUpdateRepresentative_Validation_EmptyEmail(t *testing.T) {
	a := &corpSigningAdatper{s: &mockCorpSigningService{}}
	opt := &models.RepresentativeUpdateOption{RepName: "Bob", RepEmail: ""}

	r, merr := a.UpdateRepresentative("u", "l", "s", opt)
	if merr == nil {
		t.Fatal("expected model error for empty email")
	}
	if r != nil {
		t.Errorf("result should be nil, got %+v", r)
	}
}

func TestUpdateRepresentative_Validation_InvalidEmail(t *testing.T) {
	a := &corpSigningAdatper{s: &mockCorpSigningService{}}
	opt := &models.RepresentativeUpdateOption{RepName: "Bob", RepEmail: "not-an-email"}

	r, merr := a.UpdateRepresentative("u", "l", "s", opt)
	if merr == nil {
		t.Fatal("expected model error for invalid email")
	}
	if r != nil {
		t.Errorf("result should be nil, got %+v", r)
	}
}

func TestUpdateRepresentative_NoNewAccount(t *testing.T) {
	svc := &mockCorpSigningService{created: nil}
	a := &corpSigningAdatper{s: svc}
	opt := &models.RepresentativeUpdateOption{RepName: "Bob", RepEmail: "b@corp.com"}

	r, merr := a.UpdateRepresentative("u", "l", "s", opt)
	if merr != nil {
		t.Fatalf("unexpected error: %v", merr)
	}
	if r != nil {
		t.Errorf("result should be nil when no account created, got %+v", r)
	}
	if !svc.calledRep {
		t.Error("service UpdateRepresentative should be called")
	}
	if svc.repName != "Bob" || svc.repEmail != "b@corp.com" {
		t.Errorf("args: name=%s email=%s", svc.repName, svc.repEmail)
	}
}

func TestUpdateRepresentative_NewAccountCreated(t *testing.T) {
	created := &app.ManagerDTO{
		Role:      domain.RoleAdmin,
		Name:      "Bob",
		Account:   "admin_corp.com",
		Password:  []byte("init-pw"),
		EmailAddr: "b@corp.com",
	}
	svc := &mockCorpSigningService{created: created}
	a := &corpSigningAdatper{s: svc}
	opt := &models.RepresentativeUpdateOption{RepName: "Bob", RepEmail: "b@corp.com"}

	r, merr := a.UpdateRepresentative("u", "l", "s", opt)
	if merr != nil {
		t.Fatalf("unexpected error: %v", merr)
	}
	if r == nil {
		t.Fatal("result should not be nil when account created")
	}
	if r.ID != "admin_corp.com" || r.Email != "b@corp.com" || r.Role != domain.RoleAdmin {
		t.Errorf("mapped result: %+v", r)
	}
	if string(r.Password) != "init-pw" {
		t.Errorf("password: got %s", string(r.Password))
	}
}

func TestUpdateRepresentative_ServiceError(t *testing.T) {
	wantErr := errors.New("boom")
	svc := &mockCorpSigningService{err: wantErr}
	a := &corpSigningAdatper{s: svc}
	opt := &models.RepresentativeUpdateOption{RepName: "Bob", RepEmail: "b@corp.com"}

	r, merr := a.UpdateRepresentative("u", "l", "s", opt)
	if merr == nil {
		t.Fatal("expected model error")
	}
	if r != nil {
		t.Errorf("result should be nil on error, got %+v", r)
	}
}
