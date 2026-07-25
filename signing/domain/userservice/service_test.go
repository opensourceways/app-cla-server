package userservice

import (
	"errors"
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

// ---- mock ----

type mockUserRepo struct {
	repository.User

	findByAccountResult domain.User
	findByAccountErr    error
	saveErr             error

	saved *domain.User
}

func (m *mockUserRepo) FindByAccount(linkId string, a dp.Account) (domain.User, error) {
	return m.findByAccountResult, m.findByAccountErr
}

func (m *mockUserRepo) Save(u *domain.User) error {
	m.saved = u
	return m.saveErr
}

// ---- tests ----

func TestUpdateEmailByAccount_Success(t *testing.T) {
	origEmail := dp.CreateEmailAddr("old@corp.com")
	repo := &mockUserRepo{
		findByAccountResult: domain.User{
			LinkId: "link1",
			UserBasicInfo: domain.UserBasicInfo{
				Id:        "u1",
				Account:   dp.CreateAccount("admin_corp.com"),
				EmailAddr: origEmail,
			},
		},
	}
	s := NewUserService(repo, nil, nil)

	newEmail := dp.CreateEmailAddr("new@corp.com")
	if err := s.UpdateEmailByAccount("link1", dp.CreateAccount("admin_corp.com"), newEmail); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.saved == nil {
		t.Fatal("Save should be called")
	}
	if repo.saved.EmailAddr.EmailAddr() != "new@corp.com" {
		t.Errorf("saved email: got %s, want new@corp.com", repo.saved.EmailAddr.EmailAddr())
	}
	// account 不应变
	if repo.saved.Account.Account() != "admin_corp.com" {
		t.Errorf("saved account: got %s, want admin_corp.com", repo.saved.Account.Account())
	}
}

func TestUpdateEmailByAccount_AccountNotFound(t *testing.T) {
	wantErr := errors.New("not found")
	repo := &mockUserRepo{findByAccountErr: wantErr}
	s := NewUserService(repo, nil, nil)

	err := s.UpdateEmailByAccount("link1", dp.CreateAccount("admin_corp.com"), dp.CreateEmailAddr("new@corp.com"))
	if err != wantErr {
		t.Fatalf("err: got %v, want %v", err, wantErr)
	}
	if repo.saved != nil {
		t.Error("Save should NOT be called when FindByAccount fails")
	}
}

func TestUpdateEmailByAccount_SaveError(t *testing.T) {
	wantErr := errors.New("save failed")
	repo := &mockUserRepo{
		findByAccountResult: domain.User{
			UserBasicInfo: domain.UserBasicInfo{
				Account:   dp.CreateAccount("admin_corp.com"),
				EmailAddr: dp.CreateEmailAddr("old@corp.com"),
			},
		},
		saveErr: wantErr,
	}
	s := NewUserService(repo, nil, nil)

	err := s.UpdateEmailByAccount("link1", dp.CreateAccount("admin_corp.com"), dp.CreateEmailAddr("new@corp.com"))
	if err != wantErr {
		t.Fatalf("err: got %v, want %v", err, wantErr)
	}
}
