package userservice

import (
	"os"
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func TestMain(m *testing.M) {
	dp.Init(&dp.Config{
		MaxLengthOfName:     100,
		MaxLengthOfTitle:    200,
		MaxLengthOfEmail:    200,
		MaxLengthOfAccount:  100,
		MaxLengthOfCorpName: 200,
	})
	os.Exit(m.Run())
}

// fakeUserRepo embeds repository.User so only the methods used by
// UpdateEmail need to be overridden.
type fakeUserRepo struct {
	repository.User

	findByEmailOfCorpSigningFn func(linkId, csId string, e dp.EmailAddr) (domain.User, error)
	saveFn                     func(u *domain.User) error

	findCalledWith struct {
		linkId string
		csId   string
		email  string
	}

	saveCalledWith struct {
		account string
		email   string
	}
}

func (f *fakeUserRepo) FindByEmailOfCorpSigning(linkId, csId string, e dp.EmailAddr) (domain.User, error) {
	f.findCalledWith.linkId = linkId
	f.findCalledWith.csId = csId
	f.findCalledWith.email = e.EmailAddr()
	if f.findByEmailOfCorpSigningFn != nil {
		return f.findByEmailOfCorpSigningFn(linkId, csId, e)
	}
	return domain.User{}, nil
}

func (f *fakeUserRepo) Save(u *domain.User) error {
	f.saveCalledWith.account = u.Account.Account()
	f.saveCalledWith.email = u.EmailAddr.EmailAddr()
	if f.saveFn != nil {
		return f.saveFn(u)
	}
	return nil
}

func newUserServiceForTest(repo repository.User) *userService {
	return &userService{repo: repo}
}

// ---- UpdateEmail ----

func TestUpdateEmailSameEmailNoOp(t *testing.T) {
	email := dp.CreateEmailAddr("admin@old.com")
	repo := &fakeUserRepo{}

	s := newUserServiceForTest(repo)

	err := s.UpdateEmail("link1", "cs1", "admin", email, email)
	if err != nil {
		t.Fatalf("UpdateEmail same email: unexpected error %v", err)
	}

	if repo.findCalledWith.linkId != "" {
		t.Error("UpdateEmail same email should not call FindByEmailOfCorpSigning")
	}
}

func TestUpdateEmailUsesCsIdForLookup(t *testing.T) {
	oldEmail := dp.CreateEmailAddr("admin@old.com")
	newEmail := dp.CreateEmailAddr("admin@new.com")
	found := false

	repo := &fakeUserRepo{
		findByEmailOfCorpSigningFn: func(linkId, csId string, e dp.EmailAddr) (domain.User, error) {
			found = true
			if linkId != "link1" {
				t.Errorf("linkId: got %s, want link1", linkId)
			}
			if csId != "cs1" {
				t.Errorf("csId: got %s, want cs1", csId)
			}
			if e.EmailAddr() != "admin@old.com" {
				t.Errorf("email: got %s, want admin@old.com", e.EmailAddr())
			}
			return domain.User{
				LinkId:        linkId,
				CorpSigningId: csId,
				UserBasicInfo: domain.UserBasicInfo{
					Account:   dp.CreateAccount("admin_old.com"),
					EmailAddr: oldEmail,
				},
			}, nil
		},
	}

	s := newUserServiceForTest(repo)

	if err := s.UpdateEmail("link1", "cs1", "admin", oldEmail, newEmail); err != nil {
		t.Fatalf("UpdateEmail: unexpected error %v", err)
	}

	if !found {
		t.Error("UpdateEmail should call FindByEmailOfCorpSigning")
	}
}

func TestUpdateEmailFindError(t *testing.T) {
	oldEmail := dp.CreateEmailAddr("admin@old.com")
	newEmail := dp.CreateEmailAddr("admin@new.com")

	repo := &fakeUserRepo{
		findByEmailOfCorpSigningFn: func(linkId, csId string, e dp.EmailAddr) (domain.User, error) {
			return domain.User{}, commonRepo.NewErrorResourceNotFound(nil)
		},
	}

	s := newUserServiceForTest(repo)

	if err := s.UpdateEmail("link1", "cs1", "admin", oldEmail, newEmail); err == nil {
		t.Error("UpdateEmail should propagate find error")
	}
}

func TestUpdateEmailSyncsAdminIdInAccount(t *testing.T) {
	oldEmail := dp.CreateEmailAddr("rep@old.com")
	newEmail := dp.CreateEmailAddr("rep@new.com")

	repo := &fakeUserRepo{
		findByEmailOfCorpSigningFn: func(linkId, csId string, e dp.EmailAddr) (domain.User, error) {
			return domain.User{
				UserBasicInfo: domain.UserBasicInfo{
					Account:   dp.CreateAccount("stale_admin_old.com"),
					EmailAddr: oldEmail,
				},
			}, nil
		},
	}

	s := newUserServiceForTest(repo)

	if err := s.UpdateEmail("link1", "cs1", "admin1", oldEmail, newEmail); err != nil {
		t.Fatalf("UpdateEmail: unexpected error %v", err)
	}

	// The new account must use the passed adminId ("admin1") + new domain,
	// NOT the stale adminId ("stale_admin") from the old account.
	wantAccount := "admin1_new.com"
	if repo.saveCalledWith.account != wantAccount {
		t.Errorf("account: got %s, want %s", repo.saveCalledWith.account, wantAccount)
	}

	if repo.saveCalledWith.email != "rep@new.com" {
		t.Errorf("email: got %s, want rep@new.com", repo.saveCalledWith.email)
	}
}

func TestUpdateEmailSaveError(t *testing.T) {
	oldEmail := dp.CreateEmailAddr("rep@old.com")
	newEmail := dp.CreateEmailAddr("rep@new.com")
	saveErr := commonRepo.NewErrorConcurrentUpdating(nil)

	repo := &fakeUserRepo{
		findByEmailOfCorpSigningFn: func(linkId, csId string, e dp.EmailAddr) (domain.User, error) {
			return domain.User{
				UserBasicInfo: domain.UserBasicInfo{
					Account:   dp.CreateAccount("admin_old.com"),
					EmailAddr: oldEmail,
				},
			}, nil
		},
		saveFn: func(u *domain.User) error {
			return saveErr
		},
	}

	s := newUserServiceForTest(repo)

	if err := s.UpdateEmail("link1", "cs1", "admin", oldEmail, newEmail); err == nil {
		t.Error("UpdateEmail should propagate save error")
	}
}
